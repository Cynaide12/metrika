package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"metrika/internal/config"
	"metrika/internal/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	GormDB *gorm.DB
	SqlDB  *sql.DB
}

var (
	ErrNoRows        = sql.ErrNoRows
	ErrAlreadyExists = gorm.ErrDuplicatedKey
)

func New(cfg *config.Config) (*Repository, error) {

	const fn = "internal.storage.New"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DBServer.Host,
		cfg.DBServer.Username,
		cfg.DBServer.Password,
		cfg.DBServer.DBName,
		cfg.DBServer.Port,
	)

	GormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	SqlDB, err := GormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	//миграции
	GormDB.AutoMigrate(&models.Event{}, &models.Domain{}, &models.User{}, &models.UserSession{})

	return &Repository{
		GormDB: GormDB,
		SqlDB:  SqlDB,
	}, nil
}

func (s *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{
		GormDB: tx,
		SqlDB:  s.SqlDB,
	}
}

// *EVENTS
func (s *Repository) SaveEvents(events []models.Event) error {
	var fn = "internal.repository.SaveEvent"

	tx := s.GormDB.Begin()

	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.Event{}).Create(&events).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", fn, err)
	}

	var eventIds []uint
	for _, event := range events {
		eventIds = append(eventIds, event.ID)
	}

	if err := tx.Model(&models.UserSession{}).Where("active = true AND id IN ?", eventIds).Updates(&models.UserSession{LastActive: time.Now()}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", fn, err)
	}

	return tx.Commit().Error
}

// *SESSIONS
func (s *Repository) CreateNewSession(session *models.UserSession) error {
	var fn = "internal.repository.CreateNewSession"

	if err := s.GormDB.Model(&models.UserSession{}).Create(&session).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s Repository) GetActiveSessions(sessions *[]models.UserSession, limit int) error {
	var fn = "internal.repository.GetActiveSessions"

	if err := s.GormDB.Model(&models.UserSession{}).Limit(limit).Find(sessions).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoRows
		}
		return fmt.Errorf("%s: %s", fn, err)
	}
	return nil
}

func (s Repository) GetStaleSessions(limit int) ([]models.UserSession, error) {
	var fn = "internal.repository.GetStaleSessions"

	var sessions []models.UserSession

	if err := s.GormDB.Raw(`
	SELECT s.id, s.last_active FROM user_sessions s WHERE s.active = true AND s.last_active < NOW() - INTERVAL  '25 minutes'
	AND NOT EXISTS 
	(SELECT 1 FROM events e WHERE e.session_id=s.id AND e.timestamp > NOW() - INTERVAL '30 minutes') 
	LIMIT $1
	`, limit).Scan(&sessions).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sessions, ErrNoRows
		}
		return sessions, fmt.Errorf("%s: %s", fn, err)
	}
	return sessions, nil
}

func (s *Repository) CloseSession(session_ids []uint) error{
	var fn = "internal.repository.GetStaleSessions"

	if err := s.GormDB.Exec("UPDATE user_sessions SET active = false, end_time = NOW() WHERE id = ANY($1)", session_ids).Error; err != nil{
		return fmt.Errorf("%s: %w", fn, err)
	}
	return nil
}

func (s *Repository) AddSessions(sessions *[]models.UserSession) error {
	var fn = "internal.repository.AddSessions"

	if err := s.GormDB.Model(&models.UserSession{}).Create(&sessions).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// *USERS

func (s *Repository) GetOrCreateUser(fingerprint string, domain_id uint) (models.User, error) {
	var fn = "internal.repository.GetOrCreateUser"

	if err := s.GormDB.Model(&models.UserSession{}).Where("f_id = ?", fingerprint).FirstOrCreate(&models.User{Fingerprint: fingerprint, DomainID: domain_id}).Error; err != nil {
		return models.User{}, fmt.Errorf("%s: %w", fn, err)
	}

	return models.User{}, nil
}

func (s *Repository) AddUsers(users *[]models.User) error {
	var fn = "internal.repository.AddUsers"

	if err := s.GormDB.Model(&models.User{}).Create(&users).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// *DOMAINS

type GetDomainsOptions struct {
	limit             *int
	preload_relations *bool
	site_url          *string
}

func (s *Repository) GetDomains(domains *[]models.Domain, opts GetDomainsOptions) error {
	var fn = "internal.repository.GetDomain"

	query := s.GormDB.Model(&models.Domain{})

	if opts.preload_relations != nil {
		query = query.Preload("Users")
	}

	if opts.limit != nil {
		query = query.Limit(*opts.limit)
	}

	if opts.site_url != nil {
		query = query.Where("site_url ILIKE %?%", opts.site_url)
	}

	if err := query.Find(&domains).Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) GetDomain(domain *models.Domain, domainUrl string) error {
	var fn = "internal.repository.GetDomain"

	if err := s.GormDB.Model(models.Domain{}).Where("site_url = ?", domainUrl).First(&domain).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoRows
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) AddDomain(domain *models.Domain) error {
	var fn = "internal.repository.AddDomain"

	if err := s.GormDB.Model(models.Domain{}).Create(&domain).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}

		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) GetCountDomainUsers(domainId uint) (int64, error) {
	var fn = "internal.repository.GetCountDomainUsers"

	var count int64
	if err := s.GormDB.Model(&models.User{}).Where("domain_id = ?", domainId).Count(&count).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return count, fmt.Errorf("%s: %w", fn, err)
	}

	return count, nil
}

type GetDomainUsersOptions struct {
	limit *int
}

func (s *Repository) GetDomainUsers(users *[]models.User, domainId uint, opts GetDomainUsersOptions) error {
	var fn = "internal.repository.GetCountDomainUsers"

	query := s.GormDB.Model(&models.User{}).Where("domain_id = ?", domainId)

	if opts.limit != nil {
		query = query.Limit(*opts.limit)
	}

	if err := query.Find(&users).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
