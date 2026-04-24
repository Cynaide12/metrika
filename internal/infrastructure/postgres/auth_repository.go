package postgres

import (
	"context"
	"errors"
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
	db := getDB(ctx, r.db)

	var user User
	if err := db.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.ErrUserNotFound
		}
		return nil, err
	}

	authUser := auth.User{
		ID:       user.ID,
		Email:    user.Email,
		Password: auth.NewPasswordFromHash(user.Password),
	}

	return &authUser, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, auser *auth.User) error {
	db := getDB(ctx, r.db)

	user := User{
		Email:    auser.Email,
		Password: auser.Password.Hash,
	}

	if err := db.Model(&user).Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return auth.ErrUserAlreadyExists
		}
		return err
	}

	//записываем id созданного юзера
	auser.SetID(user.ID)


	return nil
}
