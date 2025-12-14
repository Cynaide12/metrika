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
	mockGuestsIds   []uint
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

	mockDomainId, mockGuestsIds, mockSessionIds, err := m.seedMockData()
	if err != nil {
		panic("ошибка при инициализации моковых данных")
	}

	m.mockDomainId = mockDomainId
	m.mockSessionIds = mockSessionIds
	m.mockGuestsIds = mockGuestsIds

	return m
}

func (m MockService) seedMockData() (mockDomainId uint, mockGuestsIds []uint, mockSessionIds []uint, err error) {
	mockDomainUrl := "https://test.ru"

	domain := models.Domain{
		SiteURL: mockDomainUrl,
	}

	//проверяем наличие мокового домена
	err = m.repo.GetDomain(&domain, mockDomainUrl)
	if err != nil && !errors.Is(err, repository.ErrNoRows) {
		return 0, mockGuestsIds, mockSessionIds, err
	}

	if errors.Is(err, repository.ErrNoRows) {
		//добавляем моковый домен
		if err := m.repo.AddDomain(&domain); err != nil {
			return 0, mockGuestsIds, mockSessionIds, err
		}
	}

	//инициализируем моковых юзеров домена
	mockGuestsIds, err = m.initMockGuests(domain.ID)
	if err != nil {
		return 0, mockGuestsIds, mockSessionIds, err
	}

	//генерируем сессии для моковых юзеров
	var sessions []models.GuestSession
	for _, id := range mockGuestsIds {
		sessions = append(sessions, *m.generator.GenerateMockGuestSession(id))
	}

	if err := m.repo.AddSessions(&sessions); err != nil {
		return 0, mockGuestsIds, mockSessionIds, err
	}

	for _, session := range sessions {
		mockSessionIds = append(mockSessionIds, session.ID)
	}

	return domain.ID, mockGuestsIds, mockSessionIds, nil
}

func (m MockService) initMockGuests(mockDomainId uint) ([]uint, error) {
	var mockGuestsIds []uint

	IsFilledGuests, guestsToGenerate, err := m.checkLimitDomainGuests(mockDomainId)
	if err != nil {
		return mockGuestsIds, err
	}

	//если юзеры уже добавлены до максимума - не добавляем
	if !IsFilledGuests {
		var mockGuestsToAdd []models.Guest
		for range guestsToGenerate {
			//генерируем юзера
			mockGuestsToAdd = append(mockGuestsToAdd, m.generator.GenerateMockGuest(mockDomainId))
		}

		//добавляем юзеров
		if err := m.repo.AddGuests(&mockGuestsToAdd); err != nil {
			return mockGuestsIds, err
		}
	}

	var mockGuests []models.Guest
	//получаем юзеров домена
	if err := m.repo.GetDomainGuests(&mockGuests, mockDomainId, repository.GetDomainGuestsOptions{}); err != nil {
		return mockGuestsIds, err
	}

	//собираем id
	for _, guest := range mockGuests {
		mockGuestsIds = append(mockGuestsIds, guest.ID)
	}

	return mockGuestsIds, nil
}

func (m MockService) checkLimitDomainGuests(mockDomainId uint) (bool, uint, error) {
	var fn = "internal.service.mock_service.CheckLimitDomainGuests"
	logger := m.log.With("fn", fn)

	count, err := m.repo.GetCountDomainGuests(mockDomainId)
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
		case <-m.closeChan:
			return 
		}
	}
}


func (m MockService) StopEventsGenerator(){
	close(m.closeChan)
}
