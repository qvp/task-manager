package api

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"
	"task/task"
	"time"
)

type server struct {
	tasks *task.Server
}

func Run(tasks *task.Server, addr string) {
	handlers := server{
		tasks: tasks,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", handlers.createTask)
	mux.HandleFunc("GET /tasks/{id}", handlers.getTask)

	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *server) createTask(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var taskRequest TaskCreateRequest
	err := json.NewDecoder(r.Body).Decode(&taskRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskID, err := s.tasks.Add(taskRequest.Data, processTask)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := TaskCreateResponse{
		ID: taskID,
	}

	jsonResponse(w, response)
}

func (s *server) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	taskID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid parameter id", http.StatusBadRequest)
		return
	}

	taskPtr, ok := s.tasks.Get(taskID)
	if !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	response := TaskResponse{
		ID:     taskPtr.ID,
		Data:   taskPtr.Data,
		Result: taskPtr.Result,
		Status: taskPtr.Status,
	}

	jsonResponse(w, response)
}

func jsonResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
	}
}

func processTask(data string) (string, error) {
	time.Sleep(time.Second)
	result := "Done: " + data
	return result, nil
}
