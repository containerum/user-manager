package impl

import (
	"git.containerum.net/ch/json-types/errors"
	"git.containerum.net/ch/user-manager/server"
)

const (
	userNotFound            = "user was not found"
	userAlreadyExists       = "user %s is already registered"
	userNotPartiallyDeleted = "user %s is not partially deleted"
	domainInBlacklist       = "email domain %s is in blacklist"
	linkNotFound            = "link %s was not found or already used or expired"
	profilesNotFound        = "profiles not found"
	waitForResend           = "can`t resend link now, please wait %d seconds"
	oneTimeTokenNotFound    = "one-time token %s not exists or already used"
	resourceNotSupported    = "resource %s not supported now"
	activationNeeded        = "Activate your account please. Check your email"
	invalidPassword         = "invalid password provided"
	linkNotForPassword      = "link %s is not for password changing"
	linkNotForConfirm       = "link %s is not for activation"
	userBanned              = "user banned"
	tokenNotOwnedByUser     = "token %s not owned by user %s"
	adminRequired           = "you don`t have access to do this"
)

// internal errors
var (
	userGetFailed              = &server.InternalError{Err: errors.New("get user from db failed")}
	linkGetFailed              = &server.InternalError{Err: errors.New("get user link from db failed")}
	linkCreateFailed           = &server.InternalError{Err: errors.New("link creation failed")}
	emailSendFailed            = &server.InternalError{Err: errors.New("email send failed")}
	tokenCreateFailed          = &server.InternalError{Err: errors.New("token create failed")}
	getTokenFailed             = &server.InternalError{Err: errors.New("get token from db failed")}
	oauthUserInfoGetFailed     = &server.InternalError{Err: errors.New("get user info over oauth failed")}
	boundAccountsGetFailed     = &server.InternalError{Err: errors.New("get user bound accounts from db failed")}
	bindAccountFailed          = &server.InternalError{Err: errors.New("bind account failed")}
	reCaptchaRequestFailed     = &server.InternalError{Err: errors.New("reCaptcha check request failed")}
	deleteTokenFailed          = &server.InternalError{Err: errors.New("delete token failed")}
	userUpdateFailed           = &server.InternalError{Err: errors.New("update user in db failed")}
	oneTimeTokenDeleteFailed   = &server.InternalError{Err: errors.New("one-time token delete failed")}
	blacklistDomainCheckFailed = &server.InternalError{Err: errors.New("check if domain blacklisted failed")}
	userCreateFailed           = &server.InternalError{Err: errors.New("create user in db failed")}
	profileGetFailed           = &server.InternalError{Err: errors.New("get profile failed")}
	blacklistUserFailed        = &server.InternalError{Err: errors.New("user blacklisting failed")}
	blacklistUsersGetFailed    = &server.InternalError{Err: errors.New("get blacklisted users from db failed")}
	webAPIRequestFailed        = &server.InternalError{Err: errors.New("web-api request failed")}
)
