package main

import (
	"log"
	"os"
	"reflect"
	"runtime"
)

func main() {

	f := initLog(0, 0, 0) // major,patch,minor
	if f != os.Stdout {
		defer func() {
			f.Close()
		}()
	}

	log.Println("Welcome to the gate-jump server! Setting up environment...")

	log.Println("Loading Configuration")
	LoadConfig("config/config.json")

	log.Println("Initializing Database")
	s := Server{LogFile: f}
	s.Initialize(Config.Database.Username, Config.Database.Password, Config.Database.Dsn)

	log.Println("Initializing Routes")
	s.InitializeRoutes()

	log.Println("Starting the gate-jump server now! Ctrl+C to quit.")
	s.Run(Config.Port)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func MyCaller() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)
	if n == 0 {
		return "n/a"
	}
	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}
	return fun.Name()
}
