// The iso8601 package encodes and decodes time.Time in JSON in
// ISO 8601 format, without subsecond resolution or time zone info.
package iso8601

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

const Format = "2006-01-02T15:04:05"
const jsonFormat = `"` + Format + `"`

var fixedZone = time.FixedZone("", 0)

type Time struct {
	Time  time.Time
	Valid bool
}

// New constructs a new iso8601.Time instance from an existing
// time.Time instance.  This causes the nanosecond field to be set to
// 0, and its time zone set to a fixed zone with no offset from UTC
// (but it is *not* UTC itself).
func New(t time.Time) Time {
	return Time{
		Time: time.Date(
			t.Year(),
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			t.Second(),
			0,
			fixedZone,
		),
		Valid: true,
	}
}

func (it Time) MarshalJSON() ([]byte, error) {
	if !it.Valid {
		return []byte("null"), nil
	}
	return []byte(time.Time(it.Time).Format(jsonFormat)), nil
}

func (it *Time) UnmarshalJSON(data []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		t, err := time.ParseInLocation(jsonFormat, x, fixedZone)
		if err == nil {
			it = &Time{
				Time: t,
				Valid: true,
			}
		}
	case nil:
		it.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type iso8601.Time", reflect.TypeOf(v).Name())
	}
	it.Valid = err == nil
	return err
}

func (it Time) String() string {
	return time.Time(it.Time).String()
}
