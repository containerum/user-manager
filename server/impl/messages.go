package impl

import (
	"git.containerum.net/ch/json-types/errors"
	"git.containerum.net/ch/user-manager/server"
)

const (
	userAlreadyExists       = "user %s is already registered"
	userNotPartiallyDeleted = "user %s is not partially deleted"
	domainInBlacklist       = "email domain %s is in blacklist"
	linkNotFound            = "link %s was not found or already used or expired"
	waitForResend           = "can`t resend link now, please wait %d seconds"
	oneTimeTokenNotFound    = "one-time token %s not exists or already used"
	resourceNotSupported    = "resource %s not supported now"
	linkNotForPassword      = "link %s is not for password changing"
	linkNotForConfirm       = "link %s is not for activation"
	userBanned              = "user banned"
	tokenNotOwnedByUser     = "token %s not owned by user %s"
	invalidReCaptcha        = "invalid recaptcha"
	domainBlacklistEmpty    = "domain blacklist is empty"
)

// internal errors
var (
	userGetFailed    = &server.InternalError{Err: errors.New("get user from db failed")}
	userUpdateFailed = &server.InternalError{Err: errors.New("update user in db failed")}
	userCreateFailed = &server.InternalError{Err: errors.New("create user in db failed")}

	linkGetFailed    = &server.InternalError{Err: errors.New("get user link from db failed")}
	linkCreateFailed = &server.InternalError{Err: errors.New("create link in db failed")}
	linkUpdateFailed = &server.InternalError{Err: errors.New("link update failed")}

	oneTimeTokenCreateFailed = &server.InternalError{Err: errors.New("one-time token create failed")}
	oneTimeTokenGetFailed    = &server.InternalError{Err: errors.New("get one-time token from db failed")}
	oneTimeTokenDeleteFailed = &server.InternalError{Err: errors.New("one-time token delete failed")}
	oneTimeTokenUpdateFailed = &server.InternalError{Err: errors.New("update one-token in db failed")}

	oauthUserInfoGetFailed = &server.InternalError{Err: errors.New("get user info over oauth failed")}

	boundAccountsGetFailed    = &server.InternalError{Err: errors.New("get user bound accounts from db failed")}
	boundAccountsDeleteFailed = &server.InternalError{Err: errors.New("delete user bound account failed")}
	bindAccountFailed         = &server.InternalError{Err: errors.New("bind account failed")}

	reCaptchaRequestFailed = &server.InternalError{Err: errors.New("reCaptcha check request failed")}

	blacklistDomainCheckFailed = &server.InternalError{Err: errors.New("check if domain blacklisted failed")}

	profileGetFailed    = &server.InternalError{Err: errors.New("get profile failed")}
	profileUpdateFailed = &server.InternalError{Err: errors.New("update profile in db failed")}
	profileCreateFailed = &server.InternalError{Err: errors.New("create profile in db failed")}

	blacklistUserFailed     = &server.InternalError{Err: errors.New("user blacklisting failed")}
	blacklistUsersGetFailed = &server.InternalError{Err: errors.New("get blacklisted users from db failed")}

	tokenCreateFailed = &server.InternalError{Err: errors.New("token create failed")}
	tokenDeleteFailed = &server.InternalError{Err: errors.New("token delete failed")}

	resourceAccessGetFailed = &server.InternalError{Err: errors.New("resource access get failed")}

	blacklistDomainFailed   = &server.InternalError{Err: errors.New("domain blacklisting failed")}
	unblacklistDomainFailed = &server.InternalError{Err: errors.New("removing domain from blacklisting failed")}
)

var (
	userNotFound       = &server.NotFoundError{Err: errors.New("User with such email does not exist")}
	profilesNotFound   = &server.NotFoundError{Err: errors.New("profiles not found")}
	domainNotBlacklist = &server.NotFoundError{Err: errors.New("domain is not in blacklist")}
)

var (
	webApiLoginFailed = &server.AccessDeniedError{Err: errors.New("login using webapi failed")}
	adminRequired     = &server.AccessDeniedError{Err: errors.New("you don`t have access to do this")}
	invalidPassword   = &server.AccessDeniedError{Err: errors.New("invalid password provided")}
	activationNeeded  = &server.AccessDeniedError{Err: errors.New("Activate your account please. Check your email")}
)
