package scheduler

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestTask struct {
	Name string
}

func (t TestTask) Exec() {
	log.Printf("Task %s executed", t.Name)
}

func TestAdd(t *testing.T) {
	cleanupTicker = time.NewTicker(time.Second)
	now := time.Now()

	tt1 := TestTask{Name: "1"}
	tt2 := TestTask{Name: "2"}
	tt3 := TestTask{Name: "3"}
	Add(tt1, now.Add(time.Second))
	Add(tt2, now.Add(time.Second+500*time.Millisecond))
	Add(tt3, now.Add(2*time.Second))

	time.Sleep(4 * time.Second)

	assert.Empty(t, queue.tasks)
}
