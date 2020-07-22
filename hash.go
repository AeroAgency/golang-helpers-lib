package helpers

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Hash struct{}

func (h *Hash) GetHashStringByStruct(inputStruct interface{}) string {
	var b bytes.Buffer
	gob.NewEncoder(&b).Encode(inputStruct)
	return fmt.Sprintf("%x", b.Bytes())
}
