package worker

// This a simple work queue package which handles background process in a queue
import (
	"time"
)

type Task interface {
	Background() error
}

type Worker struct {
	Nworker   int
	Task      chan Task
	Taskqueue chan chan Task
	quit      chan bool
}

func Newworker(nworker int, taskqueue chan chan Task) Worker {
	return Worker{
		Nworker:   nworker,
		Task:      make(chan Task),
		Taskqueue: taskqueue,
		quit:      make(chan bool),
	}
}

func (w *Worker) Workqueue() {
	go func() {
		for {
			w.Taskqueue <- w.Task
			select {
			case work := <-w.Task:
				// retry if an error occurs when executing a background task
				for i := 0; i < 3; i++ {
					if err := work.Background(); err == nil {
						return
					}
					time.Sleep(500 * time.Millisecond)
				}
			case <-w.quit:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
