package domain

import "time"

// ==========  entities ==========

// User полная модель пользователя для хранения в БД
type User struct {
	ID           string    `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// UserPublic публичное представление модели User, без чувствительных данных
type UserPublic struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserSummary минимальное представление модели User для вложения в другие ответы
type UserSummary struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type TokenClaims struct {
    UserID    string
    ExpiresAt time.Time
}

// ==========  DTO ==========

// RegisterRequest данные для регистрации нового пользователя
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest данные для входа в систему
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateUserRequest данные для обновления профиля
type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
}

// AuthResponse ответ при успешной регистрации/логине
type AuthResponse struct {
	AccessToken string     `json:"access_token"`
	TokenType   string     `json:"token_type"`
	ExpiresIn   time.Duration  `json:"expires_in"`
	User        UserPublic `json:"user"`
}

// ========== methods ==========

// ToPublic конвертирует полную модель User в публичное представление UserPublic, скрывая пароль
func (u *User) ToPublic() UserPublic {
	return UserPublic{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
	}
}

// ToSummary возвращает минимальное представление User
func (u *User) ToSummary() UserSummary {
	return UserSummary{
		ID:       u.ID,
		Username: u.Username,
	}
}