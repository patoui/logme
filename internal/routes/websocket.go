package routes

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/patoui/logme/internal/global"
	"github.com/patoui/logme/internal/queue"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type message struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// TODO: check account ID against auth'ed user for access
func Websocket(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	// TODO: move this to the queue
	go sendLiveTail(c)
}

func sendLiveTail(c *websocket.Conn) {
	ctx := context.Background()
	ticker := time.Tick(1 * time.Second)

	for {
		select {
		case <-ticker:
			// TODO: determine if optimizations are needed
			tailLen, tailLenErr := queue.Len(global.LiveTailKey)

			if tailLenErr != nil {
				log.Fatal(tailLenErr)
				return
			}

			if tailLen == 0 {
				continue
			}

			rawLogs, nextErr := queue.Next(global.LiveTailKey, tailLen)

			if nextErr != nil {
				log.Fatal(nextErr)
				return
			}

			for _, rawLog := range rawLogs {
				// TODO: get timestamp from log
				err := wsjson.Write(ctx, c, message{
					Message:   rawLog,
					Timestamp: time.Now().Format(time.RFC3339),
				})
				if err != nil {
					log.Fatal(err)
					return
				}
			}
		}
	}
}
