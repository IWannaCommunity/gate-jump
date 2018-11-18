package log

import (
	"time"
)

//Bench reports the time it took for a function to return
func Bench(start time.Time, route, ipaddr string, code int) {
	elapsed := time.Since(start)

	Info("%s took %s and returned %s from %s", route, elapsed, code, ipaddr)
}
