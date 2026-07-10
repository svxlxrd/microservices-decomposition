package service

import (
	"bookshelf/auth-service/internal/domain"
	"bookshelf/auth-service/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameExists     = errors.New("username already exists")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewUserService(repo *repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *UserService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (s *UserService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return []byte(s.jwtSecret), nil
		})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return "", fmt.Errorf("token expired")
		}
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("invalid subject")
	}

	return sub, nil
}

func (s *UserService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// валидация
	if len(req.Username) < 3 {
		return nil, ErrInvalidUsername
	}

	if len(req.Password) < 8 {
		return nil, ErrInvalidPassword
	}

	if req.Email == "" {
		return nil, ErrInvalidEmail
	}

	// проверка уникальности
	emailExists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, ErrUserExists
	}

	usernameExists, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, ErrUsernameExists
	}

	// хеширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash: %w", err)
	}

	// создание пользователя
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// генерация токена
	token, err := s.generateToken(createdUser.ID)
	if err != nil {
		return nil, err
	}

	// генерируем токен и возвращаем ответ
	return &domain.AuthResponse{
		AccessToken: token,
		TokenType:   "access",
		ExpiresIn:   1 * time.Hour,
		User: domain.UserPublic{
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (s *UserService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	if req.Email == "" {
		return nil, ErrInvalidEmail
	}

	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: token,
		TokenType:   "access",
		ExpiresIn:   1 * time.Hour,
		User:        user.ToPublic(),
	}, nil
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, fmt.Errorf("get user: %w", err)
		}
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, userID string, req domain.UpdateUserRequest) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Username != "" {
		if len(req.Username) < 3 {
			return nil, ErrInvalidUsername
		}

		existingUser, err := s.repo.GetByUsername(ctx, req.Username)
		switch {
		case errors.Is(err, repository.ErrUserNotFound):

		case err != nil:
			return nil, err

		case existingUser.ID != userID:
			return nil, ErrUsernameExists
		}

		user.Username = req.Username
	}

	// 4. Сохраняем
	if err = s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
