package scheduler

import (
	"log"
	"slices"
	"sync"
	"time"
)

type Task interface {
	Exec()
}

type scheduledTask struct {
	executable Task
	when       time.Time
	done       bool
}

var queue = struct {
	mu    sync.Mutex
	tasks []scheduledTask
}{}

var (
	execTicker       = time.NewTicker(time.Second)
	cleanupTicker    = time.NewTicker(time.Minute)
	schedulerStarted = false
)

func Add(task Task, t time.Time) {
	t = t.UTC()
	if !t.After(time.Now().UTC()) {
		log.Printf("time has to be in future, got %v", t)
		return
	}

	if !schedulerStarted {
		startExecLoop()
		startCleanupLoop()
		schedulerStarted = true
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	queue.tasks = append(
		queue.tasks,
		scheduledTask{
			executable: task,
			when:       t,
		},
	)
}

// isTimesEqualInSeconds проверяет, что две даты равны с точностью до секунды
func isTimesEqualInSeconds(t1 time.Time, t2 time.Time) bool {
	t1Second := t1.Truncate(time.Second)
	t2Second := t2.Truncate(time.Second)

	return t1Second.Equal(t2Second)
}

func startExecLoop() {
	go func() {
		for {
			now := <-execTicker.C
			queue.mu.Lock()
			for i, task := range queue.tasks {
				if task.done || !isTimesEqualInSeconds(now.UTC(), task.when) {
					continue
				}

				task.executable.Exec()
				queue.tasks[i].done = true
			}
			queue.mu.Unlock()
		}
	}()
}

func startCleanupLoop() {
	go func() {
		for {
			<-cleanupTicker.C
			queue.mu.Lock()
			queue.tasks = slices.DeleteFunc(queue.tasks, func(t scheduledTask) bool { return t.done })
			queue.mu.Unlock()
		}
	}()
}
