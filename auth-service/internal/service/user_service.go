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

// tokens
func (s *UserService) generateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
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

func (s *UserService) createAuthResponse(user *domain.User) (*domain.AuthResponse, error) {
	token, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: token,
		TokenType:   "access",
		ExpiresIn:   24 * time.Hour,
		User:        user.ToPublic(),
	}, nil
}

// main logic
func (s *UserService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	if len(req.Username) < 3 {
		return nil, domain.ErrInvalidUsername
	}

	if len(req.Password) < 8 {
		return nil, domain.ErrInvalidPassword
	}

	if req.Email == "" {
		return nil, domain.ErrInvalidEmail
	}

	emailExists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if emailExists {
		return nil, domain.ErrUserExists
	}

	usernameExists, err := s.repo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	if usernameExists {
		return nil, domain.ErrUsernameExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
	}

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.createAuthResponse(createdUser)
}


func (s *UserService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	if req.Email == "" {
		return nil, domain.ErrInvalidEmail
	}

	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.createAuthResponse(user)
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.UserPublic, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, domain.ErrUserNotFound
		default:
			return nil, fmt.Errorf("get user: %w", err)
		}
	}

	public := user.ToPublic()

	return &public, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req domain.UpdateUserRequest) (*domain.UserPublic, error) {

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, domain.ErrUserNotFound
		default:
			return nil, fmt.Errorf("get user: %w", err)
		}
	}

	if req.Username != "" {
		if len(req.Username) < 3 {
			return nil, domain.ErrInvalidUsername
		}

		existingUser, err := s.repo.GetByUsername(ctx, req.Username)
		switch {
		case errors.Is(err, repository.ErrUserNotFound):

		case err != nil:
			return nil, err

		case existingUser.ID != userID:
			return nil, domain.ErrUsernameExists
		}

		user.Username = req.Username
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	public := user.ToPublic()

	return &public, nil
}
