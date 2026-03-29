package workers

import (
	"log"
	"sync"
	"time"
)

type Notification struct {
	OrderID string
	UserID  string
	Type    string
	Message string
}

type NotificationWorker struct {
	queue chan Notification
}

func NewNotificationWorker(bufferSize int) *NotificationWorker {
	return &NotificationWorker{
		queue: make(chan Notification, bufferSize),
	}
}

func (w *NotificationWorker) Start(n int, wg *sync.WaitGroup) {
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("Worker %d started\n", i)
			for notif := range w.queue {
				time.Sleep(50 * time.Millisecond)
				log.Printf("[NOTIFICATION WORKER %d] %s for order %s (user: %s): %s",
					i, notif.Type, notif.OrderID, notif.UserID, notif.Message)
			}
		}()
	}
}

func (w *NotificationWorker) Submit(notif Notification) bool {
	select {
	case w.queue <- notif:
		return true
	default:
		return false
	}
}

func (w *NotificationWorker) Stop() {
	close(w.queue)
}
