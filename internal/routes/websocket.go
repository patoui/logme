package routes

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/patoui/logme/internal/db"
	"github.com/patoui/logme/internal/models"
	"github.com/rueian/valkey-go"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var cache valkey.Client

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

	go sendLiveTail(c)
}

func sendLiveTail(c *websocket.Conn) {
	ctx := context.Background()
	ticker := time.Tick(1 * time.Second)

	for {
		select {
		case <-ticker:
			cache, cacheErr := db.Cache()
			if cacheErr != nil {
				log.Fatal(cacheErr)
				return
			}
			defer cache.Close()

			// TODO: determine if optimizations are needed
			tailLen, tailLenErr := cache.Do(
				ctx,
				cache.B().Llen().Key(models.LiveTailKey).Build(),
			).ToInt64()

			if tailLenErr != nil {
				log.Fatal(tailLenErr)
				return
			}

			if tailLen == 0 {
				continue
			}

			rawLogs, lPopErr := cache.Do(
				ctx,
				cache.B().Lpop().Key(models.LiveTailKey).Count(tailLen).Build(),
			).AsStrSlice()

			if lPopErr != nil {
				log.Fatal(lPopErr)
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
