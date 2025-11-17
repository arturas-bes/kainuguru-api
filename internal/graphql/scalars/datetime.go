package scalars

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// DateTime is a custom scalar type for date/time values
type DateTime time.Time

// MarshalGQL implements graphql.Marshaler for DateTime
func (d DateTime) MarshalGQL(w io.Writer) {
	t := time.Time(d)
	if t.IsZero() {
		io.WriteString(w, "null")
		return
	}
	io.WriteString(w, strconv.Quote(t.Format(time.RFC3339)))
}

// UnmarshalGQL implements graphql.Unmarshaler for DateTime
func (d *DateTime) UnmarshalGQL(v interface{}) error {
	t, err := UnmarshalDateTime(v)
	if err != nil {
		return err
	}
	*d = DateTime(t)
	return nil
}

// MarshalDateTime marshals time.Time to RFC3339 string format
func MarshalDateTime(t time.Time) graphql.Marshaler {
	if t.IsZero() {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		// Use RFC3339 format for ISO 8601 compliance
		io.WriteString(w, strconv.Quote(t.Format(time.RFC3339)))
	})
}

// UnmarshalDateTime unmarshals string to time.Time
// Accepts multiple date/time formats:
// - RFC3339: "2006-01-02T15:04:05Z07:00"
// - ISO 8601: "2006-01-02T15:04:05Z"
// - Date only: "2006-01-02"
func UnmarshalDateTime(v interface{}) (time.Time, error) {
	if v == nil {
		return time.Time{}, nil
	}

	switch v := v.(type) {
	case string:
		// Try parsing with different formats
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}

		var lastErr error
		for _, format := range formats {
			t, err := time.Parse(format, v)
			if err == nil {
				return t, nil
			}
			lastErr = err
		}

		return time.Time{}, fmt.Errorf("unable to parse datetime %q: %w", v, lastErr)

	case int, int32, int64:
		// Unix timestamp support
		var timestamp int64
		switch t := v.(type) {
		case int:
			timestamp = int64(t)
		case int32:
			timestamp = int64(t)
		case int64:
			timestamp = t
		}
		return time.Unix(timestamp, 0), nil

	default:
		return time.Time{}, errors.New("datetime must be a string or integer (unix timestamp)")
	}
}
