package models

import (
	"context"
	"encoding/json"
	"fmt"
	logger "log"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/patoui/logme/internal/db"
	"github.com/patoui/logme/internal/helpers"
)

const accountIdKey = "accountId"
const DateFormat = "2006-01-02 15:04:05"
const LiveTailKey = "live_tail:logs"

type Log struct {
	Uuid       uuid.UUID          `ch:"uuid" json:"uuid"`
	Name       string             `ch:"name" json:"name"`
	AccountId  uint32             `ch:"account_id" json:"account_id"`
	Content    string             `ch:"content" json:"content"`
	DateTime   helpers.CustomTime `ch:"dt" json:"timestamp"`
	RecordedAt time.Time          `ch:"recorded_at" json:"recorded_at"`
}

func (log *Log) Create(dbLogs driver.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logsErr := dbLogs.AsyncInsert(
		ctx,
		fmt.Sprintf(
			`INSERT INTO logs (uuid, account_id, dt, name, content, recorded_at) VALUES ('%s', %d, '%s', '%s', '%s', '%s')`,
			uuid.New().String(),
			log.AccountId,
			log.DateTime.Time.Format(DateFormat),
			log.Name,
			log.Content,
			log.RecordedAt.Format(DateFormat),
		),
		os.Getenv("DB_LOGS_ASYNC_WAIT") == "true",
	)

	if logsErr != nil {
		return logsErr
	}

	cache, cClientErr := db.Cache()
	if cClientErr != nil {
		logger.Fatal(cClientErr)
		return cClientErr
	}
	defer cache.Close()

	logJson, marshalErr := json.Marshal(log)
	if marshalErr != nil {
		logger.Fatal(marshalErr)
		return marshalErr
	}

	cacheCtx := context.Background()
	cacheErr := cache.Do(
		cacheCtx,
		cache.B().Lpush().Key(LiveTailKey).Element(string(logJson)).Build(),
	).Error()

	if cacheErr != nil {
		logger.Fatal(cacheErr)
		return cacheErr
	}

	return logsErr
}

func List(dbLogs driver.Conn, accountId int, query string) ([]Log, error) {
	rows, logsErr := dbLogs.Query(
		context.Background(),
		fmt.Sprintf("SELECT * FROM logs WHERE account_id = %d ORDER BY dt DESC", accountId),
	)

	if logsErr != nil {
		return nil, logsErr
	}

	var logs []Log
	for rows.Next() {
		var currentLog struct {
			Uuid       uuid.UUID `ch:"uuid"`
			Name       string    `ch:"name"`
			AccountId  uint32    `ch:"account_id"`
			Content    string    `ch:"content"`
			DateTime   time.Time `ch:"dt"`
			RecordedAt time.Time `ch:"recorded_at"`
		}
		if scanErr := rows.ScanStruct(&currentLog); scanErr != nil {
			return nil, scanErr
		}
		logs = append(logs, Log{
			Uuid:       currentLog.Uuid,
			Name:       currentLog.Name,
			AccountId:  currentLog.AccountId,
			Content:    currentLog.Content,
			DateTime:   helpers.CustomTime{Time: currentLog.DateTime},
			RecordedAt: currentLog.RecordedAt,
		})
	}

	return logs, nil
}
