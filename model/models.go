package model

import (
	"time"
)

// User - структура пользователя
type User struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"-"`
	Name     string `json:"name"`
}

// Note - Заметки
type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Login - Данные для входа
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register - Данные для регистрации
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Name     string `json:"name"`
}
