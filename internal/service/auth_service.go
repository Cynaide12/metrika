package service

import (
	"errors"
	"fmt"
	"log/slog"
	"metrika/internal/config"
	"metrika/internal/models"
	"metrika/internal/repository"
	lib_jwt "metrika/lib/jwt"
	"metrika/lib/logger/sl"
	lib_password "metrika/lib/password"
	"net/http"

	"gorm.io/gorm"
)

var (
	RefreshToken        = errors.New("need to generate a new access token")
	RefreshTokenInvalid = errors.New("refresh token invalid")
	UserNotFound        = errors.New("user not found")
	UserAlreadyExsits   = errors.New("user with this email already exists")
	Unauthorized        = errors.New("Unauthorized")
	UserSessionNotFound = errors.New("user session not found")
)

type AuthService struct {
	log  *slog.Logger
	repo *repository.Repository
	cfg *config.Config
}

func NewAuthService(repo *repository.Repository, log *slog.Logger, cfg *config.Config) *AuthService {
	return &AuthService{
		log,
		repo,
		cfg,
	}
}


func (s AuthService) Login(email string, password string, userAgent string) (accessToken string, refreshToken string, err error) {
	fn := "internal.services.auth.Login"

	//начинаем транзакцию
	tx := s.repo.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	txStorage := s.repo.WithTx(tx)

	//находим юзера в базе
	var user models.User
	if err := txStorage.GetUserByEmail(&user, email); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return "", "", UserNotFound
		}
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	//проверяем правильность пароля
	if ok := lib_password.CheckPasswordHash(password, user.Password); !ok {
		return "", "", Unauthorized
	}

	//создаем новую сессию юзера и jwt тоцены
	accessToken, refreshToken, err = s.CreateUserSession(tx, txStorage, user, userAgent)
	if err != nil {
		tx.Rollback()
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	return accessToken, refreshToken, nil
}

func (s AuthService) Register(user *models.User, userAgent string) (accessToken string, refreshToken string, err error) {
	fn := "internal.services.auth.Register"

	//начинаем транзакцию
	tx := s.repo.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	txStorage := s.repo.WithTx(tx)

	//сохраняем юзера в базе
	if err := txStorage.CreateUser(user); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return "", "", UserAlreadyExsits
		}

		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	//создаем новую сессию юзера и jwt тоцены
	accessToken, refreshToken, err = s.CreateUserSession(tx, txStorage, *user, userAgent)
	if err != nil {
		tx.Rollback()
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	return accessToken, refreshToken, nil

}

func (s AuthService) RefreshTokens(refreshTokenCookie *http.Cookie) (accessToken string, refreshToken string, err error) {
	fn := "internal.services.auth.RefreshTokens"

	s.log.With("fn", slog.String("fn", fn))

	//начинаем транзакцию
	tx := s.repo.GormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	txStorage := s.repo.WithTx(tx)

	refreshTokenString := refreshTokenCookie.Value

	//получаем данные юзера из токена и валидируем его
	jwtClaims, err := lib_jwt.ValidateJWT(s.cfg.JWTSecret, refreshTokenString)
	if err != nil {
		tx.Rollback()
		return "", "", RefreshTokenInvalid
	}

	//ищем сессию юзера
	var userSession models.UserSession

	if err := txStorage.GetUserSession(&userSession, jwtClaims.SessionID); err != nil {
		s.log.Error("ошибка получения сессии юзера по id из рефреш токена", sl.Err(err))
		tx.Rollback()
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	//генерируем новую пару токенов
	accessToken, refreshToken, err = lib_jwt.GeneratePairJWT(s.cfg.JWTSecret, jwtClaims.Email, jwtClaims.UserID, userSession.ID)
	if err != nil {
		s.log.Error("ошибка генерации пары jwt токенов для юзера", sl.Err(err))
		tx.Rollback()
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	//обновляем refresh токен у сессии юзера
	userSession.RefreshToken = refreshToken

	if err := txStorage.UpdateUserSession(&userSession); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return "", "", UserAlreadyExsits
		}
		tx.Rollback()

		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}

	return accessToken, refreshToken, nil

}

func (s AuthService) CreateUserSession(tx *gorm.DB, txStorage *repository.Repository, user models.User, userAgent string) (accessToken string, refreshToken string, err error) {
	fn := "internal.services.auth.CreateUserSession"

	//создаем новую сессию для юзера
	session := models.UserSession{
		UserID: user.ID,
		UserAgent: userAgent,
	}

	if err := txStorage.CreateUserSession(&session); err != nil {
		tx.Rollback()
		return "", "", fmt.Errorf("%s: ошибка при создании сессии юзера: %w", fn, err)
	}

	//генерируем jwt токены
	accessToken, refreshToken, err = lib_jwt.GeneratePairJWT(s.cfg.JWTSecret, user.Email, user.ID, session.ID)
	if err != nil {
		tx.Rollback()
		return "", "", fmt.Errorf("%s: ошибка при генерации пары токенов: %w", fn, err)
	}

	//сохраняем рефреш токен сессии в базе
	session.RefreshToken = refreshToken
	if err := txStorage.UpdateUserSession(&session); err != nil {
		tx.Rollback()
		return "", "", fmt.Errorf("%s: ошибка при обновлении сессии юзера: %w", fn, err)
	}

	return accessToken, refreshToken, nil
}

func (s AuthService) Logout(session_id uint) (err error) {
	fn := "internal.services.auth.Logout"

	//удаляем сессию юзера
	if err := s.repo.DeleteUserSession(session_id); err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return UserSessionNotFound
		}

		return fmt.Errorf("%s: %w", fn, err)
	}
	return nil

}

