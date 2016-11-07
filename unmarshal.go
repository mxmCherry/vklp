package vklp

import (
	"encoding/json"
	"fmt"
)

func unmarshal(b []byte, vs ...interface{}) error {
	r := readerPool.For(b)
	defer readerPool.Put(r)

	d := json.NewDecoder(r)

	if !d.More() {
		return nil
	}
	if t, err := d.Token(); err != nil {
		return err
	} else if delim, ok := t.(json.Delim); ok && delim != json.Delim('[') {
		return fmt.Errorf("decodeInto: expected JSON value or array opening bracket '[', got '%s'", delim.String())
	}

	for i := 0; i < len(vs); i++ {
		if !d.More() {
			return nil
		}
		if err := d.Decode(vs[i]); err != nil {
			return err
		}
	}
	return nil
}
