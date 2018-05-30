package validation

import (
	"errors"
)

const (
	isRequired      = "field %v is required"
	isRequiredSlice = "field %v is required in element %v"
)

var (
	errInvalidID = errors.New("ID should be UUID")
)
