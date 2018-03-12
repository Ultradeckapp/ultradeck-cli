package client

import (
	"fmt"
	"log"
	"os"

	"github.com/twinj/uuid"
)

func NewUUID() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}

func DebugMsg(msg string) {
	if os.Getenv("ULTRADECK_DEBUG") != "" {
		log.Println(msg)
	}
}
