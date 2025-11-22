package repository

import (
	"database/sql"
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
	GormDB.AutoMigrate()

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

// * EVENTS
func (s *Repository) SaveEvents(events []models.Event) error {
	var fn = "internal.repository.SaveEvent"

	if err := s.GormDB.Model(&models.Event{}).Create(&events).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
