package postgres

import (
	"fmt"
	"metrika/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg *config.Config) (*gorm.DB, error) {

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

	//миграции
	GormDB.AutoMigrate(&Event{}, &User{}, &Guest{}, &GuestSession{}, &UserSession{}, &Domain{}, &RecordEvent{})

	return GormDB, err
}
