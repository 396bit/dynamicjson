package dynamicjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

var unmarshaller = map[string]func(json.RawMessage, *Container) error{}
var namedUnmarshaller = map[string]string{}
var namedMarshaller = map[string]string{}

// Register any type for unmarshalling.
// Only registered types can be unmarshalled!
func Register[T any](names ...string) {
	name := typeName(reflect.TypeFor[T]())
	log.Printf("typedjson register %s", name)
	unmarshaller[name] = func(data json.RawMessage, dest *Container) error {
		var dst T
		if err := json.Unmarshal(data, &dst); err != nil {
			return err
		}
		dest.Value = dst
		return nil
	}
	for _, n := range names {
		namedMarshaller[name] = n
		namedUnmarshaller[n] = name
	}
}

// the Wrapper used in your json struct
type Container struct {
	Value any
}

func (c *Container) UnmarshalJSON(bytes []byte) error {

	helper := map[string]json.RawMessage{}

	if err := json.Unmarshal(bytes, &helper); err != nil {
		return err
	}

	switch len(helper) {
	case 0:
		c.Value = nil
		return nil
	case 1:
		for k, v := range helper {
			unmarshal, found := unmarshaller[k]
			if !found {
				var named string
				named, found = namedUnmarshaller[k]
				if found {
					unmarshal, found = unmarshaller[named]
				}
			}
			if found {
				return unmarshal(v, c)
			} else {
				return fmt.Errorf("dont know how to unmarshal (not registered): %s", k)
			}
		}
		panic("unreachable")

	default:
		return fmt.Errorf("dont know how to unmarshal (multiple keys?!)")
	}
}

func (c Container) MarshalJSON() ([]byte, error) {
	t := reflect.TypeOf(c.Value)
	if t != nil {
		name := typeName(t)
		if named, found := namedMarshaller[name]; found {
			name = named
		}
		return json.Marshal(
			map[string]any{
				name: c.Value,
			},
		)
	}
	return json.Marshal(nil)
}

func typeName(t reflect.Type) string {
	result := strings.Builder{}
	switch t.Kind() {
	case reflect.Slice:
		result.WriteString("[]")
		result.WriteString(typeName(t.Elem()))
	case reflect.Map:
		result.WriteString("map[")
		result.WriteString(typeName(t.Key()))
		result.WriteRune(']')
		result.WriteString(typeName(t.Elem()))
	case reflect.Pointer:
		result.WriteString(typeName(t.Elem()))
	default:
		if p := t.PkgPath(); p != "" {
			result.WriteString(p)
			result.WriteRune('.')
		}
		if t.Name() == "" {
			panic(errors.New("cant register unnamed type"))
		}
		result.WriteString(t.Name())
	}
	return result.String()
}
