package log

import (
	"log"
)

//Verbose logs additional info, or anything that isn't important
func Verbose(argv ...interface{}) {
	log.Printf("[VERBOSE]\t", argv)
}

//Debug logs information that could be helpful when debugging
func Debug(argv ...interface{}) {
	log.Printf("[DEBUG]\t", argv)
}

//Info logs information that could be important
func Info(argv ...interface{}) {
	log.Printf("[INFO]\t", argv)
}

//Warning logs information that points to a possible issue or point of failure
func Warning(argv ...interface{}) {
	log.Printf("[WARNING]\t", argv)
}

//Error logs information about an issue that actually occured
func Error(argv ...interface{}) {
	log.Printf("[ERROR]\t", argv)
}

//Fatal any last words?
func Fatal(argv ...interface{}) {
	log.Printf("[FATAL]\t", argv)
	panic(argv)
}

//Wtf this shouldn't happen
func Wtf(argv ...interface{}) {
	log.Printf("[WTF]\t", argv)
}
