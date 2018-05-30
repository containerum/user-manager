package validation

import (
	"errors"
)

const (
	isRequired = "field %v is required"
	notBase64  = "field %v should be encoded in base64"
	moreZero   = "field %v should be >0"
)

var (
	errInvalidID = errors.New("ID should be UUID")
)
