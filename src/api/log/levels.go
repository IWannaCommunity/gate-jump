package log

import (
	"log"
)

func Verbose(argv ...interface{}) {
	log.Println("[VERBOSE]", argv)
}

func Debug(argv ...interface{}) {
	log.Println("[DEBUG]", argv)
}

func Info(argv ...interface{}) {
	log.Println("[INFO]", argv)
}

func Warning(argv ...interface{}) {
	log.Println("[WARNING]", argv)
}

func Error(argv ...interface{}) {
	log.Println("[ERROR]", argv)
}

func Fatal(argv ...interface{}) {
	log.Println("[FATAL]", argv)
	panic(argv)
}

func Wtf(argv ...interface{}) {
	log.Println("[WTF]", argv)
}
