package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"time"

	"encoding/hex"

	"crypto/md5" // nolint: gas

	"golang.org/x/crypto/pbkdf2"
)

const pwdIteration = 30
const keyLen = 32

// GenSalt generates a salt from given data.
// It also appends current time in nanoseconds to args before salt generation.
// Salt generator is a chained sha256 (salt = sha256(salt, args[i])).
// Number of iterations is a number of args + 1.
// Result salt returned in base64-encoded string.
func GenSalt(args ...string) string {
	timeSalt := strconv.FormatInt(time.Now().UnixNano(), 10)
	args = append(args, timeSalt)

	randomByteSalt := make([]byte, 10)
	rand.Read(randomByteSalt)
	args = append(args, string(randomByteSalt))

	resultSalt := make([]byte, 0)

	for i := len(args) - 1; i >= 0; i-- { // More random data goes first
		t := sha256.Sum256(append(resultSalt, []byte(args[i])...))
		resultSalt = t[:]
	}
	return base64.StdEncoding.EncodeToString(resultSalt)
}

// GetByteKey generates a salted password using pbkdf2 algorithm.
func GetByteKey(username, pwd, salt string) []byte {
	return pbkdf2.Key([]byte(WebAPIPasswordEncode(username, pwd)), []byte(salt), pwdIteration, keyLen, sha256.New)
}

// WebAPIPasswordEncode needed to encode password with function from old web-api to allow old users to login.
func WebAPIPasswordEncode(username, plainPass string) string {
	sum := md5.Sum([]byte(username + plainPass)) // nolint: gas
	return hex.EncodeToString(sum[:])
}

// GetKey works same as GetByteKey but returns result as string.
func GetKey(username, pwd, salt string) string {
	bKey := GetByteKey(username, pwd, salt)
	return base64.StdEncoding.EncodeToString(bKey)
}

// CheckPassword allows to compare password from request with salted value from database.
func CheckPassword(username, pwd, salt, key string) bool {
	return key == GetKey(username, pwd, salt)
}
