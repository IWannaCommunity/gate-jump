package main

import (
	"log"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func httpError(w http.ResponseWriter, reason string, errorCode int) {
	if errorCode == 500 { // big boo boo
		log.Panic("Unexpected Result Occured {ERROR:", errorCode, ";REASON:", reason, "}")
	} else { // minor boo boo
		log.Println("Error Occured {ERROR:", errorCode, ";REASON:", reason, "}")
	}
	http.Error(w, reason, errorCode)
}

// Used in general response codes for simplicity sake where we don't care what happened
// but still make a habit of calling httpError/httpResponse instead of http.Error
// just the same thing as http.Error
func httpResponse(w http.ResponseWriter, reason string, responseCode int) {
	http.Error(w, reason, responseCode)
}

// httpResponseJson does a json return to the response writer
func httpResponseJson(w http.ResponseWriter, responseCode int, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	w.Write(json)
}

// Recover from panics
func Recover(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				log.Printf("Recovered from panic: %v \n %s", r, debug.Stack())
				httpError(w, http.StatusText(http.StatusInternalServerError), 500)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// NewLoggingResponseWriter helper function for passing data
func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

// Logger logs the information regarding routes in a simple way
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := NewLoggingResponseWriter(w)
		inner.ServeHTTP(lw, r)

		log.Printf(
			"%s\t%s\t%s\t%d",
			r.Method,
			r.RequestURI,
			time.Since(start),
			lw.statusCode,
		)
	})
}

// GetFunctionName helper function for returning the function's name
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
