package service

import (
	"errors"
	"fmt"
	"log/slog"
	"metrika/internal/models"
	"metrika/internal/repository"
	"metrika/internal/tracker"
	"metrika/lib/logger/sl"
)

type MockService struct {
	log                  *slog.Logger
	repo                 *repository.Repository
	tracker              *tracker.Tracker
	maxMockUsersInDomain int64
	mockDomainId         uint
}

func NewMockService(repo *repository.Repository, log *slog.Logger, tracker *tracker.Tracker, maxMockUsersInDomain int64) *MockService {
	s := &MockService{
		log:                  log,
		repo:                 repo,
		tracker:              tracker,
		maxMockUsersInDomain: maxMockUsersInDomain,
	}

	mockDomainId, err := s.init()
	if err != nil{
		panic(fmt.Sprintf("ошибка при инициализации мокового домена"))
	}

	return s
}

func (m MockService) init() (uint, error) {

	mockDomainUrl := "https://test.ru"

	domain := models.Domain{
		SiteURL: mockDomainUrl,
	}

	//проверяем наличие мокового домена
	err := m.repo.GetDomain(&domain, mockDomainUrl)
	if err != nil && errors.Is(err, repository.ErrNoRows) {
		return 0, err
	}

	if errors.Is(err, repository.ErrNoRows) {
		//добавляем моковый домен
		if err := m.repo.AddDomain(&domain); err != nil {
			return 0, err
		}
	}

	return domain.ID, nil
}

//TODO: доделать геннерацию тестовых юзеров в тестовом домене
//TODO: подумать как лучше сделать - генерацию юзеров и доменов здесь или генерацию юзеров вынести в mock пакет

func (m MockService) initMockUsers(mockDomainId uint) error {

	IsNotFilledUsers, err :=  m.checkLimitDomainUsers(mockDomainId)
	if err != nil{
		return err
	}

	//если юзеры уже добавлены до максимума - не добавляем
	if !IsNotFilledUsers{
		return nil
	}

	if

	

}

func (m MockService) checkLimitDomainUsers(mockDomainId uint) (bool, error) {
	var fn = "internal.service.mock_service.CheckLimitDomainUsers"
	logger := m.log.With("fn", fn)

	count, err := m.repo.GetCountDomainUsers(mockDomainId)
	if err != nil {
		logger.Error("ошибка при получении количества юзеров в тестовом домене", sl.Err(err))
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	if count > m.maxMockUsersInDomain {
		return false, nil
	}

	return true, nil
}

func (m MockService) addMockUser(user *models.User, mockDomainUrl string) (*models.User, error) {
	var fn = "internal.service.mock_service.AddMockUser"
	logger := m.log.With("fn", fn)

	//проверяем количество юзеров в тестовом домене
	//в тестовом домене все юзеры равнозначны тестовым юзерам
	var domain *models.Domain

	if err := m.repo.GetDomain(domain, mockDomainUrl); err != nil {
		logger.Error("ошибка при получении тестового домена", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return nil, err

}
