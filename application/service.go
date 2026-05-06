package application

//Пакет application подтягивает пакеты database, model.
import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	//1.2
	"personal-notes/database"
	"personal-notes/model"
)

type Service interface {
	//авторизация
	Register(req model.RegisterRequest) error
	Login(req model.LoginRequest) (string, error)
	Logout(token string) error
	//Заметки через CRUD
	CreateNote(userID string, title, content string) (*model.Note, error)
	GetNotes(userID string, page int) ([]model.Note, error)
	GetNote(userID, noteID string) (*model.Note, error)
	UpdateNote(userID, noteID, title, content string) error
	DeleteNote(userID, noteID string) error
}

type AuthService struct {
	repo database.Repository
	//Секретный ключ
	jwtSecret []byte
}

func NewService(repo database.Repository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

// 1.3 хранение пароля в hash
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

// 1.2 Создания токена для пользователя
func (s *AuthService) generateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) validateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("Ошибка токена")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("Ошибка претензии")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("Пользовательский id не найден в токенах")
	}
	return userID, nil
}

// 1.5 Проверка полей пользователя (Валидация)
func validateRegister(req model.RegisterRequest) error {
	if req.Login == "" {
		return errors.New("Логин не может быть пустым")
	}
	if len(req.Password) < 6 {
		return errors.New("Пароль не может содержать меньше 6 знаков")
	}
	//Пароль содержит a-z A-Z 0-9
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, c := range req.Password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
		if c >= 'a' && c <= 'z' {
			hasLower = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !(hasUpper && hasLower && hasDigit) {
		return errors.New("Пароль должен содержать A-Z, a-z, 0-9")
	}
	if req.Name == "" {
		return errors.New("Имя не может быть пустым")
	}
	return nil
}

// 2.3 Правило валидации 32 для заголовка 256 для заметки
func validateNote(title, content string) error {
	if len(title) > 32 {
		return errors.New("Заголовок не должен быть больше 32 символов")
	}
	if len(content) > 256 {
		return errors.New("Текст не должен быть больше 256 символов")
	}
	return nil
}

// 1.4 регистрация по api
func (s *AuthService) Register(req model.RegisterRequest) error {
	if err := validateRegister(req); err != nil {
		return err
	}
	existing, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		return fmt.Errorf("Ошибка ДБ: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("Логин уже существует")
	}
	//1.3 Хеширование пароля
	hashedPwd, err := hashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("Не удалось хешировать пароль: %w", err)
	}
	//создание пользователя
	user := &model.User{
		ID:       fmt.Sprintf("user_%d", time.Now().UnixNano()),
		Login:    req.Login,
		Password: hashedPwd,
		Name:     req.Name,
	}
	return s.repo.CreateUser(user)
}

// 1.1 , 1.2 Реализация метода вход по логину паролю и авторизация через ключ
func (s *AuthService) Login(req model.LoginRequest) (string, error) {
	user, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		return "", fmt.Errorf("Ошибка ДБ: %w", err)
	}
	if user == nil {
		return "", errors.New("Неверные учетные данные")
	}

	if err := checkPassword(user.Password, req.Password); err != nil {
		return "", errors.New("Неверные учетные данные")
	}
	return s.generateToken(user.ID)
}

// 1.6 выход
func (s *AuthService) Logout(token string) error {
	_, err := s.validateToken(token)
	return err
}

// 2.0 , 2.1 Создание заметки только для авторизиваных пользователей.
func (s *AuthService) CreateNote(userID, title, content string) (*model.Note, error) {
	if err := validateNote(title, content); err != nil {
		return nil, err
	}

	now := time.Now()
	note := &model.Note{
		ID:        fmt.Sprintf("note_%d", time.Now().UnixNano()),
		UserID:    userID,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateNote(note); err != nil {
		return nil, err
	}
	return note, nil
}

// 2.4 Взять заметки на 1 страницу
func (s *AuthService) GetNotes(userID string, page int) ([]model.Note, error) {
	if page < 1 {
		page = 1
	}
	return s.repo.GetNotesByUserID(userID, page)
}

// 2.2 Проверка что пользователю принадлежит заметка
func (s *AuthService) GetNote(userID, noteID string) (*model.Note, error) {
	note, err := s.repo.GetNoteByID(noteID, userID)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, errors.New("Вдоступе отказано")
	}
	return note, nil
}

// обновление заметки
func (s *AuthService) UpdateNote(userID, noteID, title, content string) error {
	if err := validateNote(title, content); err != nil {
		return err
	}
	existing, err := s.repo.GetNoteByID(noteID, userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("В доступе отказано")
	}
	updated := &model.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     title,
		Content:   content,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: time.Now(),
	}
	return s.repo.UpdateNote(updated)
}

// удаление заметки
func (s *AuthService) DeleteNote(userID, noteID string) error {
	return s.repo.DeleteNote(noteID, userID)
}

func (s *AuthService) GetUserIDFromToken(token string) (string, error) {
	return s.validateToken(token)
}
