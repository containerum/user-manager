package storages

import "errors"

var (
	// ErrInvalidToken returned if token is not valid for operation
	ErrInvalidToken = errors.New("invalid token received")

	// ErrTokenNotOwnedBySender returned if user is not token owner
	ErrTokenNotOwnedBySender = errors.New("can`t identify sender as token owner")
)
