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

func (s *UserService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	if len(req.Username) < 3 {
		return nil, ErrInvalidUsername
	}

	if len(req.Password) < 8 {
		return nil, ErrInvalidPassword
	}

	if req.Email == "" {
		return nil, ErrInvalidEmail
	}

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
		return nil, ErrInvalidEmail
	}

	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.createAuthResponse(user)
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.UserPublic, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
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
			return nil, ErrUserNotFound
		default:
			return nil, fmt.Errorf("get user: %w", err)
		}
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

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	public := user.ToPublic()

	return &public, nil
}
