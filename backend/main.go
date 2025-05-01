package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	rows, err := db.Query("SELECT id, title, description, status FROM tasks")
	if err != nil {
		log.Println("Erreur SQL:", err)
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
	setCORSHeaders(w)
	var t Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		log.Println("Erreur JSON:", err)
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, ?, ?)",
		t.Title, t.Description, "en attente")
	if err != nil {
		log.Println("Erreur SQL :", err)
		http.Error(w, "Erreur base de données", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func updateTaskStatus(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var payload struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Erreur JSON", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE tasks SET status = ? WHERE id = ?", payload.Status, id)
	if err != nil {
		log.Println("Erreur SQL :", err)
		http.Error(w, "Erreur base de données", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(mysql:3306)/tasksdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/tasks", getTasks).Methods("GET", "OPTIONS")
	r.HandleFunc("/tasks", createTask).Methods("POST", "OPTIONS")
	r.HandleFunc("/tasks/{id}", updateTaskStatus).Methods("PUT", "OPTIONS")

	log.Println("Backend en cours d'exécution sur le port 5000")
	http.ListenAndServe(":5000", r)
}
