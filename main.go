package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Name string `json:"name"`
}

var db *gorm.DB

func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.Create(&task)
	w.WriteHeader(http.StatusCreated)
}

func getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var task Task
	result := db.First(&task, taskID)
	if result.Error != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getAllTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []Task
	result := db.Find(&tasks)
	if result.Error != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var task Task
	result := db.First(&task, taskID)
	if result.Error != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.Save(&task)
	w.WriteHeader(http.StatusNoContent)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	var task Task
	result := db.First(&task, taskID)
	if result.Error != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	db.Delete(&task)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	var err error
	maxRetries := 5
	retryDelay := 5 * time.Second

	// Database connection retry loop
	for retries := 0; retries < maxRetries; retries++ {
		var dsn string = os.Getenv("POSTGRES_DSN")

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}

		log.Printf("Failed to connect to the database (retry %d/%d): %v\n", retries+1, maxRetries, err)

		time.Sleep(retryDelay)
	}

	if err != nil {
		log.Fatalf("Failed to connect to the database after %d retries: %v", maxRetries, err)
	}

	_ = db

	db.AutoMigrate(&Task{})

	r := mux.NewRouter()
	r.HandleFunc("/tasks", getAllTasks).Methods("GET")
	r.HandleFunc("/tasks", createTask).Methods("POST")
	r.HandleFunc("/tasks/{id}", getTask).Methods("GET")
	r.HandleFunc("/tasks/{id}", updateTask).Methods("PUT")
	r.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")

	http.Handle("/", r)

	port := ":8080"
	fmt.Printf("Server is running on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
