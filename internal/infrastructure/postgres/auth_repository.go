package postgres

import (
	"context"
	"errors"
	"fmt"
	"metrika/internal/domain/auth"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db}
}

func (r *AuthRepository) ByEmail(ctx context.Context, email string) (*auth.User, error) {
	const fn = "internal.repository.GetUserByEmail"
	var user User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	authUser := auth.User{
		ID:       user.ID,
		Email:    user.Email,
		Password: auth.NewPasswordFromHash(user.Password),
	}

	return &authUser, nil
}


//TODO: доделать
func (r *AuthRepository) CreateUser(ctx context.Context, user *auth.User) (error) {
	const fn = "internal.repository.CreateUser"dasd

	if err := r.db.Model().Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	authUser := auth.User{
		ID:       user.ID,
		Email:    user.Email,
		Password: auth.NewPasswordFromHash(user.Password),
	}

	return &authUser, nil
}
