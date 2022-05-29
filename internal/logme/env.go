package logme

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	chi "github.com/go-chi/chi/v5"
)

type Env struct {
	Db     driver.Conn
	Router *chi.Mux
}
