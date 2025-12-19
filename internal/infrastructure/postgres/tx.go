package postgres

import (
	"context"

	"gorm.io/gorm"
)

type TxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
	return &TxManager{db: db}
}

func (tm *TxManager) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, txKey{}, tx)
		return fn(ctx)
	})
}

func getDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}

type txKey struct{}
