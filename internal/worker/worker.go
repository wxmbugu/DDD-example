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

/*
**Basic Implementation of how we will send notification to our users email**
- Appointments Subscriber  -> will be used to send emails to user with < 24 hrs prior their appointment
- An Appointment is only approved by doctor
- When apporoved the appointment is place in redis as key value store
- Create a Schedule to fetch appointments every 15min and check if an appointment is < 24 hrs prior to scheduled time slot
- We put the appointment in a task queue && drop the record from redis
- We send emails to our users
*/

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
