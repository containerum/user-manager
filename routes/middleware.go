package routes

import (
	"net/http"

	"context"

	"git.containerum.net/ch/grpc-proto-files/utils"
	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/server"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

var hdrToKey = map[string]interface{}{
	umtypes.UserIDHeader:      server.UserIDContextKey,
	umtypes.UserAgentHeader:   server.UserAgentContextKey,
	umtypes.FingerprintHeader: server.FingerPrintContextKey,
	umtypes.SessionIDHeader:   server.SessionIDContextKey,
	umtypes.TokenIDHeader:     server.TokenIDContextKey,
	umtypes.ClientIPHeader:    server.ClientIPContextKey,
}

func requireHeaders(headers ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var notFoundHeaders []string
		for _, v := range headers {
			if ctx.GetHeader(v) == "" {
				notFoundHeaders = append(notFoundHeaders, v)
			}
		}
		if len(notFoundHeaders) > 0 {
			err := errors.Format("required headers %v was not provided", notFoundHeaders)
			ctx.Error(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, err)
		}
	}
}

func prepareContext(ctx *gin.Context) {
	for hn, ck := range hdrToKey {
		if hv := ctx.GetHeader(hn); hv != "" {
			rctx := context.WithValue(ctx.Request.Context(), ck, hv)
			ctx.Request = ctx.Request.WithContext(rctx)
		}
	}
}

func errorWithHTTPStatus(err error) (int, *errors.Error) {
	switch err.(type) {
	case *server.AccessDeniedError:
		return http.StatusForbidden, err.(*server.AccessDeniedError).Err
	case *server.NotFoundError:
		return http.StatusNotFound, err.(*server.NotFoundError).Err
	case *server.BadRequestError:
		return http.StatusBadRequest, err.(*server.BadRequestError).Err
	case *server.AlreadyExistsError:
		return http.StatusConflict, err.(*server.AlreadyExistsError).Err
	case *server.InternalError:
		return http.StatusInternalServerError, err.(*server.InternalError).Err
	case *server.WebAPIError:
		return err.(*server.WebAPIError).StatusCode, err.(*server.WebAPIError).Err
	default:
		if grpcStatus, ok := status.FromError(err); ok {
			if httpStatus, hasStatus := grpcutils.GRPCToHTTPCode[grpcStatus.Code()]; hasStatus {
				return httpStatus, errors.New(grpcStatus.Message())
			}
			return http.StatusInternalServerError, errors.New(grpcStatus.Err().Error())
		}
		return http.StatusInternalServerError, errors.New(err.Error())
	}
}

// needs role header
func requireAdminRole(ctx *gin.Context) {
	if ctx.GetHeader(umtypes.UserRoleHeader) != "admin" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.New("only admin can do this"))
	}

	err := srv.checkAdmin(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

}
