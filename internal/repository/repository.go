package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"metrika/internal/config"
	"metrika/internal/models"

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

	if err := s.GormDB.Model(&models.Event{}).Create(&events).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// *SESSIONS
func (s *Repository) CreateNewSession(session *models.UserSession) error {
	var fn = "internal.repository.CreateNewSession"

	if err := s.GormDB.Model(&models.UserSession{}).Create(&session).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) CloseUnactiveSessions(maxUnactivePeriod int) error {
	var fn = "internal.repository.CloseUnactiveSessions"

	if err := s.GormDB.Model(&models.UserSession{}).Where("updated").Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// *USERS

func (s *Repository) GetOrCreateUser(fingerprint string, domain_id uint) (models.User, error) {
	var fn = "internal.repository.GetUserByFingerprint"

	if err := s.GormDB.Model(&models.UserSession{}).Where("f_id = ?", fingerprint).FirstOrCreate(&models.User{Fingerprint: fingerprint, DomainID: domain_id}).Error; err != nil {
		return models.User{}, fmt.Errorf("%s: %w", fn, err)
	}

	return models.User{}, nil
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
