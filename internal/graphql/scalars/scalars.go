package scalars

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

// DateTime scalar implementation is in datetime.go

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