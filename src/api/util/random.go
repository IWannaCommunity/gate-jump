package util

import (
	srand "crypto/rand"
	"math/rand"
)

var letterBytes = []string{
					"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/-*[]{}()<>|?!.=_&^%$#@",
					"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"}
var letterIdxBits = []uint{7, 6} // 7 bits to represent a letter index

func CreateRandomString(n int, letterSet uint8, token chan []byte, err chan error) {
	seed, localerr := srand.Prime(srand.Reader, 256)

	// Make a byte array the size of variable `n`

	b := make([]byte, n)

	// Use Bitwise Math to create a random string fast
	if localerr == nil {
		source := rand.NewSource(seed.Int64())

		letterIdxMask := 1<<letterIdxBits[letterSet] - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax := 63 / letterIdxBits[letterSet]    // # of letter indices fitting in 63 bits

		for i, cache, remain := n-1, source.Int63(), letterIdxMax; i >= 0; {
			if remain == 0 {
				cache, remain = source.Int63(), letterIdxMax
			}

			if idx := int(cache & int64(letterIdxMask)); idx < len(letterBytes[letterSet]) {
				b[i] = letterBytes[letterSet][idx]
				i--
			}

			cache >>= letterIdxBits[letterSet]
			remain--
		}
	}

	err <- localerr
	token <- b

}
