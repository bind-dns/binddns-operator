package utils

import (
	"os"
)

func GetPodName() (name string) {
	return os.Getenv("POD_NAME")
}
