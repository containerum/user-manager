package server

import (
	"context"
)

var (
	FingerPrintContextKey = new(struct{})
	ClientIPContextKey    = new(struct{})
	UserAgentContextKey   = new(struct{})
	SessionIDContextKey   = new(struct{})
	UserIDContextKey      = new(struct{})
	TokenIDContextKey     = new(struct{})
)

func MustGetFingerprint(ctx context.Context) string {
	fp, ok := ctx.Value(FingerPrintContextKey).(string)
	if !ok {
		panic("fingerprint not found in context")
	}
	return fp
}

func MustGetClientIP(ctx context.Context) string {
	ip, ok := ctx.Value(ClientIPContextKey).(string)
	if !ok {
		panic("client ip not found in context")
	}
	return ip
}

func MustGetUserAgent(ctx context.Context) string {
	ip, ok := ctx.Value(UserAgentContextKey).(string)
	if !ok {
		panic("user agent not found in context")
	}
	return ip
}

func MustGetSessionID(ctx context.Context) string {
	sid, ok := ctx.Value(SessionIDContextKey).(string)
	if !ok {
		panic("session id not found in context")
	}
	return sid
}

func MustGetUserID(ctx context.Context) string {
	uid, ok := ctx.Value(UserIDContextKey).(string)
	if !ok {
		panic("user id not found in context")
	}
	return uid
}

func MustGetTokenID(ctx context.Context) string {
	uid, ok := ctx.Value(TokenIDContextKey).(string)
	if !ok {
		panic("token id not found in context")
	}
	return uid
}
