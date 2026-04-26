package application

import (
	"errors"
	"fmt"
	"go/token"
	"time"

	"gitgub.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	//1.2
	"personal-notes/database"
	"presonal-notes/model"
)

type Service interface{
	//авторизация
	Register(req model.RegisterRequest) error
	Login(req model.LoginRequest) (string, error)
	Logout(token string) error
	//Заметки через CRUD
	CreateNote(userID string, title, content string) (*model.Note,error)
	GetNotes(userID string, page int) ([]model.Note, error)
	GatNote(userID, noteID string) (*model.Note,error)
	UpdateNote(userID, noteID, title, content string) error
	DeleteNote(userID, noteID string) error
}

type AuthService struct {
	repo database.Repository
	jwtSecret []byte
}
func NewService( repo database.Repository, jwtSecret string) *AuthService {
	return &AuthService{
		repo: repo,
		jwtSecret: []byte(jwtSecret),
	}
}

//1.3
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
} 

func checkPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}
//1.2
func (s *AuthService) generateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodh265, jwt.MapClaims{
		"user_id": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) validateToken(tokenStr string) (string, error) {
	token, err :=jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
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
	if !ok{
		return "", errors.New("Пользовательский id не найден в токенах")
	}
	return userID, nil
}

//1.5
func validateRegister(req model.RegisterRequest) error {
	if req.Login =="" {
		return errors.New("Логин не может быть пустым")
	}
	if len(req.Password) < 6 {
		return errors.New("Пароль не может содержать меньше 6 знаков")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, c := range req.Password {
		if c >= 'A' && c <= 'Z'{
			hasUpper = true
		}
		if c >= 'a' && c <= 'z'{
			hasLower = true
		}
		if c >= '0' && c <= '9'{
			hasDigit = true
		}
	}
	if !(hasUpper && hasLower && hasDigit) {
		return errors.New("Пароль должен содержать A-Z, a-z, 0-9")
	}
		if req.Name == "" {
	return errors.New("Имя не может быть пустым")
	}
}

//2.3
func validateNote(title, content string) error {
	if len(title) > 32 {
		reurtn errors.New("Заголовок не должен быть больше 32 символов")
	}
	if len(content) > 256 {
		reurtn errors.New("Текст не должен быть больше 256 символов")
	}
	return nil
}

//1.4
func (s *AuthService) Register(req model.RegisterRequest) error {
	if err := validateRegister(req); err != nil {
		return err
	}
	existing, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		return fmt.Errorf("Ошибка ДБ: %w", err)
	}
	if existing != nil {
		return errors.New("Логин уже существует")
	}
	//1.3
	hashedPwd. err := hashPassword(req.Password)
	if err != nil {
		return errors.New("Не удалось хешировать пароль: %w", err)
	}

	user := &model.User{
		ID: fmt.Sprintf("user_%s", time.Now().UnixNano()),
		Login: req.Login,
		Password: hashedPwd,
		Name: req.Name,
	}
	return s.repo.CreateUser(user)
}
	
// 1.1 , 1.2
func (s *AuthService) Login(req model.LoginRequest) (string.error) {
	user, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		return "", fmt.Errorf("Ошибка ДБ: %w", err)
	}
	if err == nil {
		return "", errors.New("Неверные учетные данные")
	}

	if err := chackPassword(user.Password, req.Password); err != nil {
		return "", errors.New("Неверные учетные данные")
	}
	return s.generateToken(user.ID)
}

//1.6
func (s *AuthService) Logout(token string) error {
	_, err := s.validateToken(token)
	return err
}
//2.0 , 2.1
func (s *AuthService) CreateNote(userID, title, content string) (*model.Note, error) {
	if err := validateNote(title, content); err != nil {
		return nil, err
	}

	now := time.Now()
	note := &model.Note{
		ID: fmt.Sprintf("note_%s", time.now().UnixNano())
		UserId: UserID,
		Title: title,
		Content: content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateNote(note); err != nil {
		return nil, err
	}
	return note, nil
}

//2.4
func (s * AuthService) GetNotes(userID string, page int) ([]model.Note, error) {
	if page < 1 {
		page = 1 
	}
	return s.repo.getNotesByUserID(userID, page)
}

//2.2
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

func (s *AuthService) UpdateNote(userID, noteID, title, content string) error {
	if err := validateNote(title, content); err != nil {
		return err
	}
	existing, err := s.repo.GetNoteByID(noreID,userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("В доступе отказано")
	}
	updated := &model.Note{
		ID: noteID,
		userID: userID,
		Title: title,
		Content: content,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: time.Now(),
	}
	return s.repo.UpdateNote(updated)
}

func (s *AuthService) DeleteNote(userID, noteID string) error {
	return s.repo.DeleteNote(noteID, userID)
}

func (s *AuthService) GetUserIDFromToken(token string) (string,error) {
	return s.validateToken(token)
}