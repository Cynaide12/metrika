package mock

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"metrika/internal/config"
	domain "metrika/internal/domain/analytics"
	"metrika/internal/infrastructure/tracker"
	"metrika/pkg/logger/sl"
	"time"
)

type MockService struct {
	log            *slog.Logger
	adapter        MockServiceAdapter
	generator      *Generator
	tracker        *tracker.Tracker
	mockDomainId   uint
	mockGuestsIds  []uint
	mockSessionIds []uint
	mcfg           config.MockGenerator
	closeChan      chan struct{}
}

type MockServiceAdapter struct {
	// AddGuest(ctx context.Context, FingerprintID, IPAddress string, domainUrl string) error
	// AddSessions(ctx context.Context, sessions *[]domain.GuestSession) error
	// AddDomain(ctx context.Context, d domain.Domain) error
	// AddGuests(ctx context.Context, guests *[]domain.Guest) error
	// GetDomainGuests(ctx context.Context, domainId uint) ([]domain.Guest, error)
	// GetCountDomainGuests(ctx context.Context, domainid uint) (int64, error)
	// ByURL(ctx context.Context, url string) (domain.Domain, error)
	Events   domain.EventsRepository
	Guests   domain.GuestsRepository
	Sessions domain.GuestSessionRepository
	Domains  domain.DomainRepository
}

func NewMockService(adapter MockServiceAdapter, generator *Generator, log *slog.Logger, tracker *tracker.Tracker, mcfg config.MockGenerator) *MockService {
	m := &MockService{
		log:       log,
		adapter:   adapter,
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

	ctx := context.Background()

	var dom *domain.Domain

	//проверяем наличие мокового домена
	dom, err = m.adapter.Domains.ByURL(ctx, mockDomainUrl)
	if err != nil && !errors.Is(err, domain.ErrDomainNotFound) {
		return 0, mockGuestsIds, mockSessionIds, err
	}

	if errors.Is(err, domain.ErrDomainNotFound) {
		//добавляем моковый домен
		if dom, err = m.adapter.Domains.AddDomain(ctx, mockDomainUrl); err != nil {
			return 0, mockGuestsIds, mockSessionIds, err
		}
	}

	//инициализируем моковых юзеров домена
	mockGuestsIds, err = m.initMockGuests(ctx, dom.ID)
	if err != nil {
		return 0, mockGuestsIds, mockSessionIds, err
	}

	//генерируем сессии для моковых юзеров
	var sessions []domain.GuestSession
	for _, id := range mockGuestsIds {
		sessions = append(sessions, *m.generator.GenerateMockGuestSession(id))
	}

	newSessions, err := m.adapter.Sessions.CreateSessions(ctx, &sessions)
	if err != nil {
		return 0, mockGuestsIds, mockSessionIds, err
	}

	for _, session := range newSessions {
		mockSessionIds = append(mockSessionIds, session.ID)
	}

	return dom.ID, mockGuestsIds, mockSessionIds, nil
}

func (m MockService) initMockGuests(ctx context.Context, mockDomainId uint) ([]uint, error) {
	var mockGuestsIds []uint

	IsFilledGuests, guestsToGenerate, err := m.checkLimitDomainGuests(ctx, mockDomainId)
	if err != nil {
		return mockGuestsIds, err
	}

	//если юзеры уже добавлены до максимума - не добавляем
	if !IsFilledGuests {
		var mockGuestsToAdd []domain.Guest
		for range guestsToGenerate {
			//генерируем юзера
			mockGuestsToAdd = append(mockGuestsToAdd, m.generator.GenerateMockGuest(mockDomainId))
		}

		//TODO: НЕ ЗАБЫТЬ ЧТО В МАССИВ НИКАКИЕ ID НЕ ЗАПИСЫВАЮТСЯ НАДО ДОРАБОТРАТЬ ФУНКИЮ AddGuests В РЕПОЗИТОРИИ И ВАЩЕ НЕ ФАКТ ЧТО ЭТО НАДО!! ПРОВЕРИТЬ!!
		//добавляем юзеров
		if _, err := m.adapter.Guests.CreateGuests(ctx, &mockGuestsToAdd); err != nil {
			return mockGuestsIds, err
		}
	}

	//получаем юзеров домена
	mockGuests, err := m.adapter.Domains.GetDomainGuests(ctx, mockDomainId)
	if err != nil {
		return mockGuestsIds, err
	}

	//собираем id
	for _, guest := range *mockGuests {
		mockGuestsIds = append(mockGuestsIds, guest.ID)
	}

	return mockGuestsIds, nil
}

func (m MockService) checkLimitDomainGuests(ctx context.Context, mockDomainId uint) (bool, uint, error) {
	var fn = "internal.service.mock_service.CheckLimitDomainGuests"
	logger := m.log.With("fn", fn)

	count, err := m.adapter.Domains.GetCountDomainGuests(ctx, mockDomainId)
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

func (m MockService) StopEventsGenerator() {
	close(m.closeChan)
}
