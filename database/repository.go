package database

//database - подтягивает тока model
import (
	"database/sql"
	"fmt"

	//1.4
	"personal-notes/model"

	_ "modernc.org/sqlite"
)

// Интерфейс для работы с дапнными, в дальнейшем интерфейс удобно можно будет поменять на PostgreSQL.
type Repository interface {
	CreateUser(user *model.User) error
	GetUserByLogin(login string) (*model.User, error)
	GetUserByID(id string) (*model.User, error)

	CreateNote(note *model.Note) error
	GetNotesByUserID(userID string, page int) ([]model.Note, error)
	GetNoteByID(noteID, UserID string) (*model.Note, error)
	UpdateNote(note *model.Note) error
	DeleteNote(noteID, userID string) error
}

// Интерфейс для sqlite
type SQLiteRepository struct {
	db *sql.DB
}

// 3.1 выбрал бд sqlite
func NewRepository(dbpath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbpath)
	if err != nil {
		return nil, fmt.Errorf("Ошибка открытия БД: %w", err)
	}
	//1.3 создаеми таблицу, если ее нет хранимый в кэш
	err = createTablet(db)
	if err != nil {
		return nil, fmt.Errorf("Ошибка создания БД: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// Создание таблицы пользователя и заметок
func createTablet(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		login TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		name TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}
	//2.2 задаем фильтр по user_id
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS notes (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	return err
}

// Пользовательский метод логина
func (r *SQLiteRepository) CreateUser(user *model.User) error {
	//1.3
	_, err := r.db.Exec(
		"INSERT INTO users(id, login, password, name) VALUES (?, ?, ?, ?)",
		user.ID, user.Login, user.Password, user.Name,
	)
	return err
}

func (r *SQLiteRepository) GetUserByLogin(login string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(
		"SELECT id, login, password, name FROM users WHERE login = ?",
		login,
	).Scan(&user.ID, &user.Login, &user.Password, &user.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *SQLiteRepository) GetUserByID(id string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(
		"SELECT id, login, password, name FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Login, &user.Password, &user.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// Пользовательский метод заметок
func (r *SQLiteRepository) CreateNote(note *model.Note) error {
	_, err := r.db.Exec(`
	INSERT INTO notes (id, user_id, title, content, created_at, updated_at)
	VALUES(?, ?, ?, ?, ?, ?)`,
		note.ID, note.UserID, note.Title, note.Content, note.CreatedAt, note.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepository) GetNotesByUserID(userID string, page int) ([]model.Note, error) {
	//2.4 задается метод пагинации 25 заметок на 1 страницу.
	offset := (page - 1) * 25
	limit := 25

	rows, err := r.db.Query(
		`SELECT id, user_id, title, content, created_at, updated_at
	FROM notes
	WHERE user_id = ?
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []model.Note
	for rows.Next() {
		var n model.Note
		err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *SQLiteRepository) GetNoteByID(noteID, userID string) (*model.Note, error) {
	//2.2 проверка метод "дейстивтельно ли заметка принадлежит пользователю"
	note := &model.Note{}
	err := r.db.QueryRow(
		`SELECT id, user_id, title, content, created_at, updated_at
		FROM notes
		WHERE id = ? AND user_id = ?`,
		noteID, userID,
	).Scan(&note.ID, &note.UserID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return note, err
}

// Если заметка принадлежит пользователю. Значит обновляем
func (r *SQLiteRepository) UpdateNote(note *model.Note) error {
	_, err := r.db.Exec(
		`UPDATE notes
		SET title = ?, content = ?, updated_at = ?
		WHERE id = ? AND user_id = ?`,
		note.Title, note.Content, note.UpdatedAt, note.ID, note.UserID,
	)
	return err
}

// удаление заметки
func (r *SQLiteRepository) DeleteNote(NoteID, userID string) error {
	_, err := r.db.Exec(
		`DELETE FROM notes WHERE id = ? AND user_id = ?`,
		NoteID, userID,
	)
	return err
}
