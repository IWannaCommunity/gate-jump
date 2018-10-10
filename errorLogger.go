package main

import (
	"log"
	"net/http"
)

func httpError(w http.ResponseWriter, reason string, errorCode int) {
	if errorCode == 500 { // big boo boo
		http.Error(w, reason, errorCode)
		log.Fatal("Unexpected Result Occured {ERROR:", errorCode, ";REASON:", reason)
	}
	// booboo minor
	http.Error(w, reason, errorCode)
	log.Println("Error Occured {ERROR:")
}

// Used in general response codes for simplicity sake where we don't care what happened
// but still make a habit of calling httpError/httpResponse instead of http.Error
// just the same thing as http.Error
func httpResponse(w http.ResponseWriter, reason string, errorCode int) {
	http.Error(w, reason, errorCode)
}
