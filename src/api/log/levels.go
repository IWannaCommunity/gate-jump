package log

import (
	"log"
)

func Verbose(argv ...interface{}) {
	log.Println("[VERBOSE]\t", argv)
}

func Debug(argv ...interface{}) {
	log.Println("[DEBUG]\t", argv)
}

func Info(argv ...interface{}) {
	log.Println("[INFO]\t", argv)
}

func Warning(argv ...interface{}) {
	log.Println("[WARNING]\t", argv)
}

func Error(argv ...interface{}) {
	log.Println("[ERROR]\t", argv)
}

func Fatal(argv ...interface{}) {
	log.Println("[FATAL]\t", argv)
	panic(argv)
}

func Wtf(argv ...interface{}) {
	log.Println("[WTF]\t", argv)
}
