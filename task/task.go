package task

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
	"sync"
	"time"
)

type Server struct {
	queue      chan uuid.UUID
	links      map[uuid.UUID]*Task
	maxWorkers int64
	mu         sync.RWMutex
	wg         sync.WaitGroup
}

type Status string
type Processor func(string) (string, error)

const (
	Pending Status = "PENDING"
	Started Status = "STARTED"
	Done    Status = "DONE"
	Fail    Status = "FAIL"
)

type Task struct {
	ID        uuid.UUID
	Data      string
	Processor Processor
	Status    Status
	Result    string
}

func New(maxWorkers int64) *Server {
	if maxWorkers == 0 {
		maxWorkers = 1
	}

	return &Server{
		queue:      make(chan uuid.UUID),
		links:      make(map[uuid.UUID]*Task),
		maxWorkers: maxWorkers,
	}
}

func (s *Server) Add(data string, processor Processor) (uuid.UUID, error) {
	taskID, err := uuid.NewUUID()
	if err != nil {
		return uuid.UUID{}, err
	}

	task := Task{
		ID:        taskID,
		Data:      data,
		Processor: processor,
		Status:    Pending,
	}

	s.mu.Lock()
	s.links[taskID] = &task
	s.mu.Unlock()

	s.queue <- taskID
	fmt.Println("Add: ", data)
	return taskID, nil
}

func (s *Server) Get(taskID uuid.UUID) (*Task, bool) {
	s.mu.RLock()
	taskPtr, ok := s.links[taskID]
	s.mu.RUnlock()

	return taskPtr, ok
}

func (s *Server) Run() {
	ctx := context.TODO()
	sem := semaphore.NewWeighted(s.maxWorkers)

	for taskID := range s.queue {
		taskPtr, ok := s.Get(taskID)
		if !ok {
			fmt.Println("task removed: ", taskID)
			continue
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			fmt.Println("sem err: ", err)
		}

		s.wg.Add(1)
		go s.runTask(taskPtr, sem)
	}

	time.Sleep(time.Minute)
	s.wg.Wait()
	close(s.queue)
}

func (s *Server) Wait() {
	s.wg.Wait()
	close(s.queue)
}

func (s *Server) runTask(taskPtr *Task, sem *semaphore.Weighted) {
	defer sem.Release(1)
	defer s.wg.Done()

	taskPtr.Status = Started
	result, err := taskPtr.Processor(taskPtr.Data)

	if err != nil {
		taskPtr.Status = Fail
		taskPtr.Result = fmt.Sprint(err)
	} else {
		taskPtr.Status = Done
		taskPtr.Result = result
	}
}
