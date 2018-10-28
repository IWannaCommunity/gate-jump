package log

import (
    "log"
    "time"
    "os"
    "io"
)

func Init() *os.File {
    // Make logs directory
    if _, err := os.Stat("logs"); os.IsNotExist(err) {
        os.Mkdir("logs", os.ModePerm)
    }

    // Open a write-only file
    f, err := os.OpenFile("logs/"+time.Now().Format("2006-01-02")+".log", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0600)

    if err != nil {
        log.Fatalf("unable to open file for writing: %v", err)
    }

    log.SetOutput(io.MultiWriter(os.Stdout, f))

    return f
}
