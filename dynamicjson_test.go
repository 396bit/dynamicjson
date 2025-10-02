package dynamicjson

import (
	"encoding/json"
	"errors"
	"log"
	"testing"
)

func TestContainer(t *testing.T) {

	type x struct {
		Name string
		Age  int
	}

	Register[string]()
	Register[x]()

	type outer struct {
		Code    string
		Dynamic Container
	}

	marshalUnmarshal(t, Container{Value: nil})

	marshalUnmarshal(t, Container{Value: "justtext"})

	o := marshalUnmarshal(t, outer{})
	if o.Dynamic.Value != nil {
		t.Error(errors.New("expected error"))
	}

	o = marshalUnmarshal(t, outer{Code: "Dieter", Dynamic: Container{Value: x{Name: "Horst", Age: 42}}})
	switch val := o.Dynamic.Value.(type) {
	case x:
		log.Printf("decoded: %+v", val)
	case string:
		log.Printf("decoded string: %s", val)
	default:
		log.Fatalf("unexpected value: %T", o.Dynamic.Value)
	}

}

func marshalUnmarshal[T any](t *testing.T, v T) T {
	log.Printf("in   %T: %+v", v, v)
	out, err := json.Marshal(v)
	if err != nil {
		t.Error(err)
	}
	log.Printf("json %s", string(out))

	err = json.Unmarshal(out, &v)
	if err != nil {
		t.Error(err)
	}
	log.Printf("out  %T: %+v", v, v)
	log.Print("")
	return v
}
