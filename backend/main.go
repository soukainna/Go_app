package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

var jwtKey = []byte("secret_key")
var db *sql.DB

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// Middleware JWT
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
			http.Error(w, "Token manquant ou invalide", http.StatusUnauthorized)
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token invalide", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	rows, err := db.Query("SELECT id, title, description, status FROM tasks")
	if err != nil {
		http.Error(w, "Erreur base de données", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status)
		tasks = append(tasks, t)
	}
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var t Task
	json.NewDecoder(r.Body).Decode(&t)

	_, err := db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, ?, ?)", t.Title, t.Description, "en attente")
	if err != nil {
		http.Error(w, "Erreur d'insertion", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var t Task
	json.NewDecoder(r.Body).Decode(&t)

	_, err := db.Exec("UPDATE tasks SET status = ? WHERE id = ?", t.Status, id)
	if err != nil {
		http.Error(w, "Erreur de mise à jour", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(t)
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(mysql:3306)/taskdb")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.Use(corsMiddleware)
	r.Handle("/tasks", authMiddleware(http.HandlerFunc(getTasks))).Methods("GET", "OPTIONS")
	r.Handle("/tasks", authMiddleware(http.HandlerFunc(createTask))).Methods("POST", "OPTIONS")
	r.Handle("/tasks/{id}", authMiddleware(http.HandlerFunc(updateTask))).Methods("PUT", "OPTIONS")

	fmt.Println("Serveur backend sur port 5000")
	log.Fatal(http.ListenAndServe(":5000", r))
}
