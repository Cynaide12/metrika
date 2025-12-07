package service

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"metrika/internal/config"
	"metrika/internal/mock"
	"metrika/internal/models"
	"metrika/internal/repository"
	"metrika/internal/tracker"
	"metrika/lib/logger/sl"
	"time"
)

type MockService struct {
	log            *slog.Logger
	repo           *repository.Repository
	generator      *mock.Generator
	tracker        *tracker.Tracker
	mockDomainId   uint
	mockUsersIds   []uint
	mockSessionIds []uint
	mcfg           config.MockGenerator
	closeChan      chan struct{}
}

func NewMockService(repo *repository.Repository, generator *mock.Generator, log *slog.Logger, tracker *tracker.Tracker, mcfg config.MockGenerator) *MockService {
	m := &MockService{
		log:       log,
		repo:      repo,
		generator: generator,
		tracker:   tracker,
		mcfg:      mcfg,
		closeChan: make(chan struct{}),
	}

	mockDomainId, mockUsersIds, mockSessionIds, err := m.seedMockData()
	if err != nil {
		panic(fmt.Sprintf("ошибка при инициализации моковых данных"))
	}

	m.mockDomainId = mockDomainId
	m.mockSessionIds = mockSessionIds
	m.mockUsersIds = mockUsersIds

	return m
}

func (m MockService) seedMockData() (mockDomainId uint, mockUsersIds []uint, mockSessionIds []uint, err error) {
	mockDomainUrl := "https://test.ru"

	domain := models.Domain{
		SiteURL: mockDomainUrl,
	}

	//проверяем наличие мокового домена
	err = m.repo.GetDomain(&domain, mockDomainUrl)
	if err != nil && !errors.Is(err, repository.ErrNoRows) {
		return 0, mockUsersIds, mockSessionIds, err
	}

	if errors.Is(err, repository.ErrNoRows) {
		//добавляем моковый домен
		if err := m.repo.AddDomain(&domain); err != nil {
			return 0, mockUsersIds, mockSessionIds, err
		}
	}

	//инициализируем моковых юзеров домена
	mockUsersIds, err = m.initMockUsers(domain.ID)
	if err != nil {
		return 0, mockUsersIds, mockSessionIds, err
	}

	//генерируем сессии для моковых юзеров
	var sessions []models.UserSession
	for _, id := range mockUsersIds {
		sessions = append(sessions, *m.generator.GenerateMockUserSession(id))
	}

	if err := m.repo.AddSessions(&sessions); err != nil {
		return 0, mockUsersIds, mockSessionIds, err
	}

	for _, session := range sessions {
		mockSessionIds = append(mockSessionIds, session.ID)
	}

	return domain.ID, mockUsersIds, mockSessionIds, nil
}

func (m MockService) initMockUsers(mockDomainId uint) ([]uint, error) {
	var mockUsersIds []uint

	IsFilledUsers, usersToGenerate, err := m.checkLimitDomainUsers(mockDomainId)
	if err != nil {
		return mockUsersIds, err
	}

	//если юзеры уже добавлены до максимума - не добавляем
	if !IsFilledUsers {
		var mockUsersToAdd []models.User
		for range usersToGenerate {
			//генерируем юзера
			mockUsersToAdd = append(mockUsersToAdd, m.generator.GenerateMockUser(mockDomainId))
		}

		//добавляем юзеров
		if err := m.repo.AddUsers(&mockUsersToAdd); err != nil {
			return mockUsersIds, err
		}
	}

	var mockUsers []models.User
	//получаем юзеров домена
	if err := m.repo.GetDomainUsers(&mockUsers, mockDomainId, repository.GetDomainUsersOptions{}); err != nil {
		return mockUsersIds, err
	}

	//собираем id
	for _, user := range mockUsers {
		mockUsersIds = append(mockUsersIds, user.ID)
	}

	return mockUsersIds, nil
}

func (m MockService) checkLimitDomainUsers(mockDomainId uint) (bool, uint, error) {
	var fn = "internal.service.mock_service.CheckLimitDomainUsers"
	logger := m.log.With("fn", fn)

	count, err := m.repo.GetCountDomainUsers(mockDomainId)
	if err != nil {
		logger.Error("ошибка при получении количества юзеров в тестовом домене", sl.Err(err))
		return false, 0, fmt.Errorf("%s: %w", fn, err)
	}

	if count >= m.mcfg.MaxMockUsersInDomain {
		return true, 0, nil
	}

	return false, uint(m.mcfg.MaxMockUsersInDomain - count), nil
}

func (m MockService) StartEventsGenerator() {
	ticker := time.NewTicker(time.Second * time.Duration(m.mcfg.RandWindowSecond))

	for {
		select {
		case <-ticker.C:
			//генерируем размер пачки
			bucketSize := m.generator.GenerateBucketSize(m.mcfg.MinEventInLoop, m.mcfg.MaxEventInLoop)
			for range bucketSize {
				//отбираем случайный id сессии
				sessionId := rand.Intn(int(m.mcfg.MaxMockUsersInDomain))
				//генерируем ивент
				event := m.generator.GenerateMockEvent(uint(sessionId))

				//отправляем на обработку
				go m.tracker.TrackEvent(*event)
			}
		}
	}
}
