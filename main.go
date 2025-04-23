package main

import (
	"task/api"
	"task/task"
)

func main() {
	tasks := task.New(10)
	defer tasks.Wait()
	go tasks.Run()

	api.Run(tasks, ":8080")
}
