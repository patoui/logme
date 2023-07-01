package models

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/mitchellh/mapstructure"

	"github.com/patoui/logme/internal/helpers"
)

const accountIdKey = "accountId"
const layout = "2006-01-02 15:04:05"

type Log struct {
	Uuid      *uuid.UUID `mapstructure:"uuid" json:"uuid"`
	Name      string     `mapstructure:"name" json:"name"`
	AccountId int        `mapstructure:"account_id" json:"account_id"`
	DateTime  helpers.CustomTime `mapstructure:"timestamp" json:"timestamp"`
	Content   string     `mapstructure:"content" json:"content"`
}

type CreateLog struct {
	AccountId int        `json:"account_id"`
	Name      string     `json:"name"`
	Timestamp helpers.CustomTime `json:"timestamp"`
	Content   string     `json:"content"`
}

func (log *CreateLog) Create(db *meilisearch.Client) (error) {
    index, err := index(db)
    if err != nil {
    	return err
    }

    id := uuid.New()
    documents := []map[string]interface{}{
        {
            "uuid":       id.String(),
            "account_id": log.AccountId,
            "name":       log.Name,
            "timestamp":  log.Timestamp.Time.Format("2006-01-02 15:04:05"),
            "content":    log.Content,
        },
    }

    _, docErr := index.AddDocuments(documents, "uuid")

    return docErr
}

func List(db *meilisearch.Client, accountId int, query string) ([]Log, error) {
    index, err := index(db)
    if err != nil {
    	return nil, err
    }

    resp, searchErr := index.Search(query, &meilisearch.SearchRequest{
        Filter: fmt.Sprintf("account_id = %d", accountId),
    })

    if searchErr != nil {
        return nil, searchErr
    }

    logs, mapErr := decodeLogs(resp.Hits)
    if mapErr != nil {
        return nil, mapErr
    }

    return logs, nil
}

func index(db *meilisearch.Client) (*meilisearch.Index, error) {
	indexName := os.Getenv("PRIMARY_INDEX")
    index := db.Index(indexName)

    _, err := index.UpdateFilterableAttributes(&[]string{"account_id"})
    if err != nil {
        return nil, err
    }

    return index, nil
}

func decodeLogs(input interface{}) ([]Log, error) {
    var logs []Log

	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			stringToUUIDHookFunc(),
			stringToCustomTimeHookFunc(),
		),
		Result: &logs,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, err
	}

	decodingErr := decoder.Decode(input)
	if decodingErr != nil {
		return nil, decodingErr
	}

	return logs, nil
}

func stringToUUIDHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(uuid.UUID{}) {
			return data, nil
		}

		return uuid.Parse(data.(string))
	}
}

func stringToCustomTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(helpers.CustomTime{}) {
			return data, nil
		}

		tm, _ := time.Parse(layout, data.(string))

		return helpers.CustomTime{Time: tm}, nil
	}
}
