package vklp

import "fmt"

type Error uint8

func (e Error) Error() string {
	return fmt.Sprintf("vklp: error %d", e)
}
