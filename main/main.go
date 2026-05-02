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

	// Корень: иначе GET / даёт 404 (остальные маршруты только под /api/...).
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html><html lang="ru"><head><meta charset="utf-8"><title>Личные заметки — API</title></head><body>
<h1>API работает</h1>
<p>Это бэкенд без отдельной HTML-страницы. Доступные пути (префикс <code>/api</code>):</p>
<ul>
<li><code>POST /api/auth/register</code></li>
<li><code>POST /api/auth/login</code></li>
<li><code>POST /api/auth/logout</code></li>
<li><code>GET /api/notes</code> (заголовок авторизации как в коде)</li>
<li><code>POST /api/notes/create</code></li>
<li><code>GET|PUT|DELETE /api/notes/{id}</code></li>
</ul>
</body></html>`)
	})

	//3.0
	fmt.Println("Сервер запущен http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
