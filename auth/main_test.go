package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func setupTestDB() *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/authdb")
	if err != nil {
		log.Fatalf("Erreur connexion DB: %v", err)
	}
	return db
}

func TestLoginHandler(t *testing.T) {
	db = setupTestDB() // initialisation de la DB globale

	// Préparation de la requête
	payload := map[string]string{
		"email":    "testuser@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Login)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusUnauthorized {
		t.Errorf("Status attendu 200 ou 401, obtenu : %d", rr.Code)
	}
}
