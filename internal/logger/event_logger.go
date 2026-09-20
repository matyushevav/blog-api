package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	// eventsBufferSize - размер буфера канала событий
	eventsBufferSize = 100

	// writeDelay - задержка перед записью, демонстрирует отложенную обработку
	writeDelay = 100 * time.Millisecond

	// timestampLayout - формат метки времени в файле логов
	timestampLayout = "2006-01-02 15:04:05"
)

// EventLogger асинхронно пишет события в файл через горутину-worker.
type EventLogger struct {
	eventsChan chan string
	done       chan struct{}
	file       *os.File
}

// NewEventLogger открывает файл логов и инициализирует каналы.
func NewEventLogger(filePath string) *EventLogger {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("Failed to open events log file %q: %v", filePath, err)

		file = nil
	}

	return &EventLogger{
		eventsChan: make(chan string, eventsBufferSize),
		done:       make(chan struct{}),
		file:       file,
	}
}

// LogEvent отправляет событие в очередь.
func (el *EventLogger) LogEvent(event string) {
	select {
	case el.eventsChan <- event:
	default:
		log.Printf("Events buffer is full, event dropped: %s", event)
	}
}

// Start запускает фоновую обработку событий
func (el *EventLogger) Start() {
	go el.worker()
}

// worker читает события из канала до его закрытия.
func (el *EventLogger) worker() {
	// close(done) выполняется последним и разблокирует Stop
	defer close(el.done)

	if el.file != nil {
		defer el.file.Close()
	}

	for event := range el.eventsChan {
		time.Sleep(writeDelay)

		line := fmt.Sprintf("[%s] %s\n", time.Now().Format(timestampLayout), event)

		if el.file == nil {
			log.Print(line)

			continue
		}

		if _, err := el.file.WriteString(line); err != nil {
			log.Printf("Failed to write event: %v", err)
		}
	}
}

// Stop закрывает очередь и ждет, пока worker допишет оставшиеся события.
func (el *EventLogger) Stop() {
	close(el.eventsChan)
	<-el.done

	log.Println("Event logger stopped")
}
