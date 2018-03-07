package validation

import (
	"errors"
)

const (
	isRequired = "Field %v is required"
	notBase64  = "Field %v should be encoded in base64"
	moreZero   = "Field %v should be >0"
)

var (
	errInvalidID = errors.New("ID should be UUID")
)
