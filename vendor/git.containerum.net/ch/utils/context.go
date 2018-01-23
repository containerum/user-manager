package utils

import "context"

// Keys to inject data to context
var (
	FingerPrintContextKey = new(int)
	ClientIPContextKey    = new(int)
	UserAgentContextKey   = new(int)
	SessionIDContextKey   = new(int)
	UserIDContextKey      = new(int)
	TokenIDContextKey     = new(int)
)

// MustGetFingerprint attempts to extract client fingerprint using FingerPrintContextKey from context.
// It panics if value was not found.
func MustGetFingerprint(ctx context.Context) string {
	fp, ok := ctx.Value(FingerPrintContextKey).(string)
	if !ok {
		panic("fingerprint not found in context")
	}
	return fp
}

// MustGetClientIP attempts to extract client IP address using ClientIPContextKey from context.
// It panics if value was not found.
func MustGetClientIP(ctx context.Context) string {
	ip, ok := ctx.Value(ClientIPContextKey).(string)
	if !ok {
		panic("client ip not found in context")
	}
	return ip
}

// MustGetUserAgent attempts to extract client IP address using UserAgentContextKey from context.
// It panics if value was not found.
func MustGetUserAgent(ctx context.Context) string {
	ip, ok := ctx.Value(UserAgentContextKey).(string)
	if !ok {
		panic("user agent not found in context")
	}
	return ip
}

// MustGetSessionID attempts to extract session ID using SessionIDContextKey from context.
// It panics if value was not found in context.
func MustGetSessionID(ctx context.Context) string {
	sid, ok := ctx.Value(SessionIDContextKey).(string)
	if !ok {
		panic("session id not found in context")
	}
	return sid
}

// MustGetUserID attempts to extract user ID using SessionIDContextKey from context.
// It panics if value was not found in context.
func MustGetUserID(ctx context.Context) string {
	uid, ok := ctx.Value(UserIDContextKey).(string)
	if !ok {
		panic("user id not found in context")
	}
	return uid
}

// MustGetTokenID attempts to extract token ID using TokenIDContextKey from context.
// It panics if value was not found in context.
func MustGetTokenID(ctx context.Context) string {
	uid, ok := ctx.Value(TokenIDContextKey).(string)
	if !ok {
		panic("token id not found in context")
	}
	return uid
}
