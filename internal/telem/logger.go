package telem

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[voltvar] ", log.LstdFlags|log.Lmicroseconds|log.LUTC)

func Setup() {
	// placeholder (wire metrics/exporters here)
}

func Log() *log.Logger { return logger }
