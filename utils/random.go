package utils

import (
	"crypto/rand"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890" // 62 possibilities
	letterIdxBits = 6                                                                // 6 bits to represent 64 possibilities / indexes
	letterIdxMask = 1<<letterIdxBits - 1                                             // All 1-bits, as many as letterIdxBits
)

//SecureRandomString securly generates random string
func SecureRandomString(length int) (ret string, err error) {
	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes, err = SecureRandomBytes(bufferSize)
		}
		if err != nil {
			return "", err
		}
		if idx := int(randomBytes[j%length] & letterIdxMask); idx < len(letterBytes) {
			result[i] = letterBytes[idx]
			i++
		}
	}
	return string(result), nil
}

// SecureRandomBytes returns the requested number of bytes using crypto/rand
func SecureRandomBytes(length int) ([]byte, error) {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}
