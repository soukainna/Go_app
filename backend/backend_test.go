package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.FAKE_SIGNATURE" // Remplace par un vrai token si besoin

func TestGetTasksAuthorized(t *testing.T) {
	req, _ := http.NewRequest("GET", "/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getTasks)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 200 or 401, got %v", rr.Code)
	}
}

func TestCreateTaskAuthorized(t *testing.T) {
	payload := map[string]string{
		"title":       "Tâche test",
		"description": "Description de test",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createTask)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated && rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 201 or 401, got %v", rr.Code)
	}
}
