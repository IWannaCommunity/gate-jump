package util

import (
	"reflect"
	"runtime"
	"strconv"
)

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

func IsNumeric(s string) bool { // https://stackoverflow.com/a/45686455
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func IsValidPassword(s string) bool { // https://stackoverflow.com/a/25837334

	if len(s) < 8 {
		return false
	}

	numberCount := 0
	upperCount := 0
	specialCount := 0
	letterCount := 0

	for _, s := range s {
		switch {
		case unicode.IsNumber(s): // number
			numberCount++
		case unicode.IsUpper(s): // uppercase
			upperCount++
		case unicode.IsPunct(s) || unicode.IsSymbol(s): // special
			specialCount++
		case unicode.IsLetter(s) || s == ' ': // letter
			letterCount++
		default: // decline anything else
			return false
		}
	}
	return true
}
	}
