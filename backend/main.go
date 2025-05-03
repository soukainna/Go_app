package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

var db *sql.DB
var jwtKey = []byte("mon-jwt-secret-2025")

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(mysql:3306)/taskdb")
	if err != nil {
		log.Fatal("Erreur de connexion à la base de données :", err)
	}

	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.Handle("/tasks", authMiddleware(http.HandlerFunc(getTasks))).Methods("GET", "OPTIONS")
	r.Handle("/tasks", authMiddleware(http.HandlerFunc(createTask))).Methods("POST", "OPTIONS")
	r.Handle("/tasks/{id}", authMiddleware(http.HandlerFunc(updateTask))).Methods("PUT", "OPTIONS")

	log.Println("Serveur backend sur port 5000")
	http.ListenAndServe(":5000", r)
}

func getUserIDFromToken(r *http.Request) (int, error) {
	tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("token invalide")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return 0, errors.New("user_id manquant dans le token")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("format user_id invalide")
	}

	return int(userIDFloat), nil
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Token invalide", http.StatusUnauthorized)
		return
	}

	rows, err := db.Query("SELECT id, title, description, status FROM tasks WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Erreur base de données", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status); err != nil {
			http.Error(w, "Erreur lecture", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Token invalide", http.StatusUnauthorized)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO tasks (title, description, status, user_id) VALUES (?, ?, 'en attente', ?)",
		task.Title, task.Description, userID)
	if err != nil {
		http.Error(w, "Erreur insertion", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Token invalide", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}

	res, err := db.Exec("UPDATE tasks SET status = ? WHERE id = ? AND user_id = ?", task.Status, id, userID)
	if err != nil {
		http.Error(w, "Erreur mise à jour", http.StatusInternalServerError)
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		http.Error(w, "Tâche non trouvée ou non autorisée", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "Token manquant ou invalide", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
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
