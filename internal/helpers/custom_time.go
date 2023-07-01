package helpers

import (
    "fmt"
    "strings"
    "time"
)

const layout = "2006-01-02 15:04:05"

type CustomTime struct {
    Time time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
    s := strings.Trim(string(b), "\"")
    if s == "null" {
        ct.Time = time.Time{}
        return
    }
    ct.Time, err = time.Parse(layout, s)
    return
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
    if ct.Time.UnixNano() == nilTime {
        return []byte("null"), nil
    }
    return []byte(fmt.Sprintf("\"%s\"", ct.Time.Format(layout))), nil
}

var nilTime = (time.Time{}).UnixNano()

func (ct *CustomTime) IsSet() bool {
    return ct.Time.UnixNano() != nilTime
}