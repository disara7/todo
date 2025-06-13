package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-sql-driver/mysql"
	"github.com/thedevsaddam/renderer"
)

var rnd *renderer.Render
var db *sql.DB

const (
	port = ":19000"
)

type todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

func init() {
	rnd = renderer.New()

	// MySQL config
	cfg := mysql.NewConfig()
	cfg.User = "root"
	cfg.Passwd = "password@1234"
	cfg.DBName = "todo"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:55000"

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN()) // connect to MySQL
	checkErr(err)

	// Validate connection
	checkErr(db.Ping())
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := template.ParseFiles("static/home.tpl")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, completed, created_at FROM todos")
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to fetch todos",
			"error":   err,
		})
		return
	}
	defer rows.Close()

	var todoList []todo

	for rows.Next() {
		var t todo
		var completedInt int
		if err := rows.Scan(&t.ID, &t.Title, &completedInt, &t.CreatedAt); err != nil {
			rnd.JSON(w, http.StatusProcessing, renderer.M{
				"message": "Error scanning row",
				"error":   err,
			})
			return
		}
		t.Completed = completedInt == 1
		todoList = append(todoList, t)
	}
	if err := rows.Err(); err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Error retrieving rows",
			"error":   err,
		})
		return
	}
	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title is required",
		})
		return
	}

	t.CreatedAt = time.Now()
	t.Completed = false

	res, err := db.Exec("INSERT INTO todos (title, completed, created_at) VALUES (?, ?, ?)", t.Title, t.Completed, t.CreatedAt)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to insert todo",
			"error":   err,
		})
		return
	}

	id, err := res.LastInsertId()
	checkErr(err)

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"todo_id": id,
	})
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is required",
		})
		return
	}

	_, err := db.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?", t.Title, t.Completed, id)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to update todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
	})
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	_, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to delete todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
	})
}

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", homeHandler)

	r.Route("/todo", func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Server listening on port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server.")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully stopped.")
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
