package main

import (
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const folderPath = "logs/"

func initLog(major int, patch int, minor int) *os.File {
	f := os.Stdout
	if !(len(os.Args) > 1 && os.Args[1] == "debug") { // skip if in debug mode
		var err error
		fileLog := strings.Join([]string{folderPath, strconv.Itoa(int(major)), ".", strconv.Itoa(int(patch)), ".", strconv.Itoa(int(minor)), ".log"}, "")
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			os.Mkdir(folderPath, os.ModePerm)
		}
		f, err = os.OpenFile(fileLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Panic("Error Opening Log File: ", err)
		}

		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}
	return f
}
