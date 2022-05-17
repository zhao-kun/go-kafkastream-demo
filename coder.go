package main

import (
	"encoding/json"

	goka "github.com/lovoo/goka"
	"github.com/pkg/errors"
)

type coder struct{}

func (c *coder) Encode(value interface{}) ([]byte, error) {
	label, ok := value.(*Label)
	if !ok {
		return nil, errors.Errorf("value isn't map of the quota")
	}
	return json.Marshal(label)
}

func (c *coder) Decode(data []byte) (interface{}, error) {

	if len(data) == 0 {
		return nil, nil
	}

	label := Label{}
	err := json.Unmarshal(data, &label)
	return &label, err
}

func newCoder() goka.Codec {
	return &coder{}
}
