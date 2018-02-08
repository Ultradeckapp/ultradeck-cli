package client

import (
	"fmt"
	"log"

	"github.com/twinj/uuid"
)

func NewUUID() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}

func DebugMsg(msg string) {
	if 1 == 0 {
		log.Println(msg)
	}
}
