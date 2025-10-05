package base

import (
	"context"

	"template-golang/pkg/apperror"
	"template-golang/pkg/redisx"

	"gorm.io/gorm"
)

// BaseService menyediakan akses ke DB & transaksi
type BaseService struct {
	Db    *gorm.DB
	Redis *redisx.Client
}

func NewBaseService(db *gorm.DB, redis *redisx.Client) *BaseService {
	return &BaseService{
		Db:    db,
		Redis: redis,
	}
}

// DB returns the GORM DB instance
func (b *BaseService) DB() *gorm.DB {
	return b.Db
}

// InTx runs function inside transaction (with return)
func (b *BaseService) InTx(ctx context.Context, fn func(*gorm.DB) (any, error)) (any, error) {
	tx := b.Db.Begin()
	if tx.Error != nil {
		return nil, apperror.New("base_service", "in_tx", 500, tx.Error.Error(), "failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	result, err := fn(tx)
	if err != nil {
		tx.Rollback()
		return nil, apperror.New("base_service", "in_tx", 500, err, "failed to execute transaction")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apperror.New("base_service", "in_tx", 500, err, "failed to commit transaction")
	}

	return result, nil
}

// InTxVoid runs function inside transaction (no return)
func (b *BaseService) InTxVoid(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := b.Db.Begin()
	if tx.Error != nil {
		return apperror.New("base_service", "in_tx_void", 500, tx.Error, "failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return apperror.New("base_service", "in_tx_void", 500, err, "failed to execute transaction")
	}

	if err := tx.Commit().Error; err != nil {
		return apperror.New("base_service", "in_tx_void", 500, err, "failed to commit transaction")
	}

	return nil
}
