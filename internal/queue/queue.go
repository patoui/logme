package queue

import (
	"context"
	"log"

	"github.com/patoui/logme/internal/db"
	"github.com/patoui/logme/internal/global"
	"github.com/patoui/logme/internal/listener"
	"github.com/rueian/valkey-go"
)

var Listeners = map[string][]Listener{
	"log.created": {
		listener.LogCreated{},
	},
}

type Listener interface {
	Handle(data string)
}

func cache() valkey.Client {
	cache, cClientErr := db.Cache()
	if cClientErr != nil {
		log.Fatal(cClientErr)
		panic("Unable to retrieve cache client")
	}

	return cache
}

func Add(queue, data string) error {
	cache := cache()
	defer cache.Close()

	ctx := context.Background()
	cacheErr := cache.Do(
		ctx,
		cache.B().Lpush().Key(queue).Element(data).Build(),
	).Error()

	if cacheErr != nil {
		return cacheErr
	}

	return nil
}

func Next(queue string, count int64) ([]string, error) {
	cache := cache()
	defer cache.Close()

	ctx := context.Background()
	rawLogs, lPopErr := cache.Do(
		ctx,
		cache.B().Rpop().Key(global.LiveTailKey).Count(count).Build(),
	).AsStrSlice()

	if lPopErr != nil {
		log.Fatal(lPopErr)
		return []string{}, lPopErr
	}

	if len(rawLogs) == 0 {
		return []string{}, nil
	}

	return rawLogs, nil
}

func Len(queue string) (int64, error) {
	cache := cache()
	defer cache.Close()

	ctx := context.Background()
	len, err := cache.Do(
		ctx,
		cache.B().Llen().Key(global.LiveTailKey).Build(),
	).ToInt64()

	if err != nil {
		log.Fatal(err)
		return 0, err
	}

	return len, nil
}
