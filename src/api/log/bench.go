package log

import (
	"time"
)

//Bench reports the time it took for a function to return
func Bench(route string, start time.Time) {
	elapsed := time.Since(start)

	Info("%s took %s", route, elapsed)
}
