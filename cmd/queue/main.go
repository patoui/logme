package main

import (
	"log"
	"time"

	"github.com/patoui/logme/internal/helper"
	iqueue "github.com/patoui/logme/internal/queue"
)

func main() {
	log.Println("STARTING QUEUE...")
	helper.LoadEnv("/home/.env")
	for {
		for event, listeners := range iqueue.Listeners {
			// TODO: determine if optimizations are needed
			queueLen, queueLenErr := iqueue.Len(event)

			if queueLenErr != nil {
				log.Fatal(queueLenErr)
				return
			}

			if queueLen == 0 {
				pause()
				continue
			}

			rawLogs, nextErr := iqueue.Next(event, queueLen)

			if nextErr != nil {
				log.Fatal(nextErr)
				return
			}

			for _, rawLog := range rawLogs {
				for _, listener := range listeners {
					go listener.Handle(rawLog)
				}
			}
		}
		pause()
	}
}

func pause() {
	time.Sleep(100 * time.Millisecond)
}
