package qlog

import (
	"log"
	"os"
)

func InitLog() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
}
