package utils

import (
	"log"
	"os"
)

// GetHostName returns machin hostname
func GetHostName() string {
	name, err := os.Hostname()
	if err != nil {
		log.Println(err)
	}

	return name
}
