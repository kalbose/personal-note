package main

import (
	"fmt"
	"log"
	"net/http"
	"personal-notes/application"
	"personal-notes/database"
	"personal-notes/transport"
)

func main() {
	repo, err := database.NewRepository("notes.db")
	if err != nil {
		log.Fatalf("База данных не запущена: %v", err)
	}
	fmt.Println("База данных установлена")

	service := application.NewService(repo, "Ключ JWT смени меня")
	fmt.Println("Служба приложения установлена")

	handler := transport.NewHandler(service)

	//1.2, 2.4
	http.HandleFunc("/api/auth/register", handler.RegisterHandler)
	http.HandleFunc("/api/auth/login", handler.LoginHandler)
	http.HandleFunc("/api/auth/logout", handler.LogoutHandler)
	//2.1
	http.HandleFunc("/api/notes", handler.AuthMiddleware(handler.GetNotesHandler))
	http.HandleFunc("/api/notes/create", handler.AuthMiddleware(handler.CreateNoteHandler))

	http.HandleFunc("/api/notes/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.AuthMiddleware(handler.GetNoteHandler)(w, r)
		case http.MethodPut:
			handler.AuthMiddleware(handler.UpdateNoteHandler)(w, r)
		case http.MethodDelete:
			handler.AuthMiddleware(handler.DeleteNoteHandler)(w, r)
		default:
			transport.WriteError(w, http.StatusMethodNotAllowed, "Метод запрещен")
		}
	})
	//3.0
	fmt.Println("Сервер запущен http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
