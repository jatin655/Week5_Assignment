package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"` 
}

var (
	tasks  []Task
	nextID = 1
	mu     sync.Mutex 
)

func main() {
	http.HandleFunc("/tasks", tasksHandler)      
	http.HandleFunc("/tasks/", taskHandler)       

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAllTasks(w, r)
	case http.MethodPost:
		createTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/tasks/"):]

	switch r.Method {
	case http.MethodGet:
		getTaskByID(w, r, idStr)
	case http.MethodPut:
		updateTask(w, r, idStr)
	case http.MethodDelete:
		deleteTask(w, r, idStr)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAllTasks(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task.ID = nextID
	nextID++
	task.Status = "pending" 
	tasks = append(tasks, task)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func getTaskByID(w http.ResponseWriter, r *http.Request, idStr string) {
	mu.Lock()
	defer mu.Unlock()
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	for _, task := range tasks {
		if task.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
			return
		}
	}
	http.NotFound(w, r)
}

func updateTask(w http.ResponseWriter, r *http.Request, idStr string) {
	mu.Lock()
	defer mu.Unlock()
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var updatedTask Task
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			tasks[i].Title = updatedTask.Title
			tasks[i].Description = updatedTask.Description
			tasks[i].Status = updatedTask.Status
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tasks[i])
			return
		}
	}
	http.NotFound(w, r)
}

func deleteTask(w http.ResponseWriter, r *http.Request, idStr string) {
	mu.Lock()
	defer mu.Unlock()
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.NotFound(w, r)
}
