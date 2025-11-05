package scalars

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

// MarshalDateTime marshals a time.Time to a GraphQL DateTime scalar
func MarshalDateTime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(t.Format(time.RFC3339)))
	})
}

// UnmarshalDateTime unmarshals a GraphQL DateTime scalar to a time.Time
func UnmarshalDateTime(v interface{}) (time.Time, error) {
	switch v := v.(type) {
	case string:
		return time.Parse(time.RFC3339, v)
	case int:
		return time.Unix(int64(v), 0), nil
	case int64:
		return time.Unix(v, 0), nil
	default:
		return time.Time{}, fmt.Errorf("cannot unmarshal %T into time.Time", v)
	}
}

// MarshalUUID marshals a uuid.UUID to a GraphQL UUID scalar
func MarshalUUID(u uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(u.String()))
	})
}

// UnmarshalUUID unmarshals a GraphQL UUID scalar to a uuid.UUID
func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	switch v := v.(type) {
	case string:
		return uuid.Parse(v)
	default:
		return uuid.UUID{}, fmt.Errorf("cannot unmarshal %T into uuid.UUID", v)
	}
}

// MarshalJSON marshals an interface{} to a GraphQL JSON scalar
func MarshalJSON(v interface{}) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		data, _ := json.Marshal(v)
		_, _ = w.Write(data)
	})
}

// UnmarshalJSON unmarshals a GraphQL JSON scalar to an interface{}
func UnmarshalJSON(v interface{}) (interface{}, error) {
	switch v := v.(type) {
	case string:
		var result interface{}
		err := json.Unmarshal([]byte(v), &result)
		return result, err
	case map[string]interface{}:
		return v, nil
	case []interface{}:
		return v, nil
	default:
		return v, nil
	}
}