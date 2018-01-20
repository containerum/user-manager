package utils

import (
	"context"
	"net/http"
	"net/textproto"

	"github.com/gin-gonic/gin"
)

var headersKey = new(int)

// SaveHeaders is a gin middleware which saves headers to request context
func SaveHeaders(ctx *gin.Context) {
	rctx := context.WithValue(ctx.Request.Context(), headersKey, ctx.Request.Header)
	ctx.Request.WithContext(rctx)
}

// RequestHeadersMap extracts saved headers from context as map[string]string (useful for resty).
// saveHeaders middleware required for operation.
func RequestHeadersMap(ctx context.Context) map[string]string {
	ret := make(map[string]string)
	for k, v := range ctx.Value(headersKey).(http.Header) {
		if len(v) > 0 {
			ret[textproto.CanonicalMIMEHeaderKey(k)] = v[0] // this is how MIMEHeader.Get() works actually
		}
	}
}

// RequestHeaders extracts saved headers from context.
// saveHeaders middleware required for operation.
func RequestHeaders(ctx context.Context) http.Header {
	return ctx.Value(headersKey).(http.Header)
}
