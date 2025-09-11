package storage

import (
	"database/sql"
	"fmt"
	"go_template/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	GormDB *gorm.DB
	SqlDB  *sql.DB
}

func New(cfg *config.Config) (*Storage, error) {

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
	GormDB.AutoMigrate(

	)

	return &Storage{
		GormDB: GormDB,
		SqlDB:  SqlDB,
	}, nil
}

func (s *Storage) WithTx(tx *gorm.DB) *Storage {
	return &Storage{
		GormDB: tx,
		SqlDB:  s.SqlDB,
	}
}
