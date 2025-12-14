package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	GormDB.AutoMigrate(&models.Event{}, &models.Domain{}, &models.Guest{}, &models.GuestSession{})

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
		eventIds = append(eventIds, event.SessionID)
	}

	if err := tx.Model(&models.GuestSession{}).Where("active = true AND id IN ?", eventIds).Updates(&models.GuestSession{LastActive: time.Now()}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", fn, err)
	}

	return tx.Commit().Error
}

// *SESSIONS
func (s *Repository) CreateNewSession(session *models.GuestSession) error {
	var fn = "internal.repository.CreateNewSession"

	if err := s.GormDB.Model(&models.GuestSession{}).Create(&session).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s Repository) GetActiveSessions(sessions *[]models.GuestSession, limit int) error {
	var fn = "internal.repository.GetActiveSessions"

	if err := s.GormDB.Model(&models.GuestSession{}).Limit(limit).Find(sessions).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoRows
		}
		return fmt.Errorf("%s: %s", fn, err)
	}
	return nil
}

func (s Repository) GetStaleSessions(limit int) ([]models.GuestSession, error) {
	var fn = "internal.repository.GetStaleSessions"

	var sessions []models.GuestSession

	if err := s.GormDB.Raw(`
	SELECT s.id, s.last_active FROM guest_sessions s WHERE s.active = true AND s.last_active < NOW() - INTERVAL  '25 minutes'
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

func (s *Repository) CloseSession(session_ids []uint) error {
	var fn = "internal.repository.GetStaleSessions"

	if err := s.GormDB.Exec("UPDATE guest_sessions SET active = false, end_time = NOW() WHERE id = ANY($1)", session_ids).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	return nil
}

func (s *Repository) AddSessions(sessions *[]models.GuestSession) error {
	var fn = "internal.repository.AddSessions"

	if err := s.GormDB.Model(&models.GuestSession{}).Create(&sessions).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// *GUESTS

func (s *Repository) GetOrCreateGuest(fingerprint string, domain_id uint) (models.Guest, error) {
	var fn = "internal.repository.GetOrCreateGuest"

	guest := models.Guest{Fingerprint: fingerprint, DomainID: domain_id}

	if err := s.GormDB.Model(&models.Guest{}).Where("f_id = ?", fingerprint).FirstOrCreate(&guest).Error; err != nil {
		return models.Guest{}, fmt.Errorf("%s: %w", fn, err)
	}

	log.Println("USER", guest.ID)

	return guest, nil
}

func (s *Repository) AddGuests(guests *[]models.Guest) error {
	var fn = "internal.repository.AddGuests"

	if err := s.GormDB.Model(&models.Guest{}).Create(&guests).Error; err != nil {
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
		query = query.Preload("Guests")
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

func (s *Repository) GetCountDomainGuests(domainId uint) (int64, error) {
	var fn = "internal.repository.GetCountDomainGuests"

	var count int64
	if err := s.GormDB.Model(&models.Guest{}).Where("domain_id = ?", domainId).Count(&count).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return count, fmt.Errorf("%s: %w", fn, err)
	}

	return count, nil
}

type GetDomainGuestsOptions struct {
	limit *int
}

func (s *Repository) GetDomainGuests(guests *[]models.Guest, domainId uint, opts GetDomainGuestsOptions) error {
	var fn = "internal.repository.GetCountDomainGuests"

	query := s.GormDB.Model(&models.Guest{}).Where("domain_id = ?", domainId)

	if opts.limit != nil {
		query = query.Limit(*opts.limit)
	}

	if err := query.Find(&guests).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// *INFO

//TODO: доделать
func (s *Repository) GetCountActiveSessions(domain_id uint) (int64, error) {
	var fn = "internal.repository.getCountActiveSessions"

	var count int64

	if err := s.GormDB.Debug().Model(&models.GuestSession{}).Exec("SELECT * FROM guest_sessions s LEFT JOIN guests u ON u.id=s.guest_id WHERE s.active = true AND u.domain_id=?", domain_id).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return count, nil
}



// *USERS
func (s *Repository) GetUserByEmail(user *models.User, email string) error {
	const fn = "internal.repository.GetUserByEmail"
	if err := s.GormDB.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoRows
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) GetUserByID(user *models.User, id uint) error {
	const fn = "internal.repository.GetUserByID"
	if err := s.GormDB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoRows
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) CreateUser(user *models.User) error {
	const fn = "internal.repository.CreateUser"
	if err := s.GormDB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrNoRows
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}



//* AUTH

func (s *Repository) CreateUserSession(session *models.UserSession) error {
	const fn = "internal.repository.CreateUserSession"
	if err := s.GormDB.Create(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) UpdateUserSession(session *models.UserSession) error {
	const fn = "internal.repository.UpdateUserSession"
	if err := s.GormDB.Where("id = ?", session.ID).Updates(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) DeleteUserSession(session_id uint) error {
	const fn = "internal.repository.DeleteUserSession"
	if err := s.GormDB.Where("id = ?", session_id).Delete(&models.UserSession{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoRows
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Repository) GetUserSession(session *models.UserSession, id uint) error {
	const fn = "storage.GetUserSession"
	if err := s.GormDB.First(&session, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoRows
		}
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}