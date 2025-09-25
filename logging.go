package gtfsrtsiri

import (
	"log"
	"os"
)

func InitLogging() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}
