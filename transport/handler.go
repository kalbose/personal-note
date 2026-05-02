package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"personal-notes/application"
	"personal-notes/model"
	"strings"
)

type Handler struct {
	service application.Service
}

func NewHandler(svc application.Service) *Handler {
	return &Handler{service: svc}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func WriteError(w http.ResponseWriter, status int, message string) {
	writeError(w, status, message)
}

func getTokenFromHeader(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		parts := strings.SplitN(strings.TrimSpace(auth), " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}
	// Старый вариант (кириллические имена заголовков)
	auth = r.Header.Get("Авторизован")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(strings.TrimSpace(auth), " ", 2)
	if len(parts) != 2 || parts[0] != "Носитель" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := getTokenFromHeader(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "Авторизация провалена")
			return
		}
		if svc, ok := h.service.(*application.AuthService); ok {
			userID, err := svc.GetUserIDFromToken(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "Ошибка токенв")
				return
			}
			ctx := context.WithValue(r.Context(), "userID", userID)
			next(w, r.WithContext(ctx))
		} else {
			writeError(w, http.StatusInternalServerError, "Ошибка сервиса")
			return
		}
	}
}

// 1.4
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		return
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Недопустимый текст")
		return
	}

	if err := h.service.Register(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Пользователь зарегестрирован"})
}

// 1.2
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Недопустимый текст")
		return
	}

	token, err := h.service.Login(req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"Токен": token})
}

// 1.6
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		return
	}

	token := getTokenFromHeader(r)
	if err := h.service.Logout(token); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Выход из системы"})
}

// 2.0, 2.1, 2.4
func (h *Handler) CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		return
	}
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "Пользователь не авторизован")
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	note, err := h.service.CreateNote(userID, req.Title, req.Content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

// 2.4
func (h *Handler) GetNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "Пользователь не авторизован")
		return
	}

	page := 1
	if p := r.URL.Query().Get("Страница"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}

	notes, err := h.service.GetNotes(userID, page)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, notes)
}

func (h *Handler) GetNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "Пользователь не авторизован")
		return
	}

	noteID := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if noteID == "" {
		writeError(w, http.StatusBadRequest, "Требуется note_id")
		return
	}

	note, err := h.service.GetNote(userID, noteID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (h *Handler) UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "Пользователь не авторизован")
		return
	}

	noteID := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if noteID == "" {
		writeError(w, http.StatusBadRequest, "Требуется note_id")
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Недопустимый текст")
		return
	}

	if err := h.service.UpdateNote(userID, noteID, req.Title, req.Content); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Запись обновлена"})
}

func (h *Handler) DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "Пользователь не авторизован")
		return
	}

	noteID := strings.TrimPrefix(r.URL.Path, "/api/notes/")
	if noteID == "" {
		writeError(w, http.StatusBadRequest, "Требуется note_id")
		return
	}

	if err := h.service.DeleteNote(userID, noteID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Запись удалена"})
}
