package impl

import (
	"git.containerum.net/ch/json-types/errors"
	"git.containerum.net/ch/user-manager/server"
)

const (
	userAlreadyExists       = "User %s is already registered"
	userNotPartiallyDeleted = "User %s is not partially deleted"
	domainInBlacklist       = "Email domain %s is in blacklist"
	linkNotFound            = "Link %s was not found or already used or expired"
	waitForResend           = "Can`t resend link now, please wait %d seconds"
	oneTimeTokenNotFound    = "One-time token %s not exists or already used" //nolint: gas
	resourceNotSupported    = "Resource %s not supported now"
	linkNotForPassword      = "Link %s is not for password changing" //nolint: gas
	linkNotForConfirm       = "Link %s is not for activation"
	userBanned              = "User banned"
	tokenNotOwnedByUser     = "Token %s not owned by user %s"
	invalidReCaptcha        = "Invalid recaptcha"
	domainBlacklistEmpty    = "Domain blacklist is empty"
)

// internal errors
var (
	userGetFailed    = &server.InternalError{Err: errors.New("Get user from db failed")}
	userUpdateFailed = &server.InternalError{Err: errors.New("Update user in db failed")}
	userCreateFailed = &server.InternalError{Err: errors.New("Create user in db failed")}

	linkGetFailed    = &server.InternalError{Err: errors.New("Get user link from db failed")}
	linkCreateFailed = &server.InternalError{Err: errors.New("Create link in db failed")}
	linkUpdateFailed = &server.InternalError{Err: errors.New("Link update failed")}

	oneTimeTokenCreateFailed = &server.InternalError{Err: errors.New("One-time token create failed")}
	oneTimeTokenGetFailed    = &server.InternalError{Err: errors.New("Get one-time token from db failed")}
	oneTimeTokenDeleteFailed = &server.InternalError{Err: errors.New("One-time token delete failed")}
	oneTimeTokenUpdateFailed = &server.InternalError{Err: errors.New("Update one-token in db failed")}

	oauthUserInfoGetFailed = &server.InternalError{Err: errors.New("Get user info over oauth failed")}

	boundAccountsGetFailed    = &server.InternalError{Err: errors.New("Get user bound accounts from db failed")}
	boundAccountsDeleteFailed = &server.InternalError{Err: errors.New("Delete user bound account failed")}
	bindAccountFailed         = &server.InternalError{Err: errors.New("Bind account failed")}

	reCaptchaRequestFailed = &server.InternalError{Err: errors.New("ReCaptcha check request failed")}

	blacklistDomainCheckFailed = &server.InternalError{Err: errors.New("Check if domain blacklisted failed")}

	profileGetFailed    = &server.InternalError{Err: errors.New("Get profile failed")}
	profileUpdateFailed = &server.InternalError{Err: errors.New("Update profile in db failed")}
	profileCreateFailed = &server.InternalError{Err: errors.New("Create profile in db failed")}

	blacklistUserFailed     = &server.InternalError{Err: errors.New("User blacklisting failed")}
	blacklistUsersGetFailed = &server.InternalError{Err: errors.New("Get blacklisted users from db failed")}

	tokenCreateFailed = &server.InternalError{Err: errors.New("Token create failed")}
	tokenDeleteFailed = &server.InternalError{Err: errors.New("Token delete failed")}

	resourceAccessGetFailed = &server.InternalError{Err: errors.New("Resource access get failed")}

	blacklistDomainFailed   = &server.InternalError{Err: errors.New("Domain blacklisting failed")}
	unblacklistDomainFailed = &server.InternalError{Err: errors.New("Removing domain from blacklisting failed")}
)

var (
	userNotFound       = &server.NotFoundError{Err: errors.New("User with such credentials does not exist")}
	profilesNotFound   = &server.NotFoundError{Err: errors.New("Profiles not found")}
	domainNotBlacklist = &server.NotFoundError{Err: errors.New("Domain is not in blacklist")}
)

var (
	webAPILoginFailed = &server.AccessDeniedError{Err: errors.New("Login using WebAPI failed")}
	adminRequired     = &server.AccessDeniedError{Err: errors.New("You don`t have access to do this")}
	invalidPassword   = &server.AccessDeniedError{Err: errors.New("Invalid password provided")}
	activationNeeded  = &server.AccessDeniedError{Err: errors.New("Activate your account please. Check your email")}
	userNotBlacklist  = &server.AccessDeniedError{Err: errors.New("User is not in blacklist")}
	oauthLoginFailed  = &server.AccessDeniedError{Err: errors.New("Invalid OAuth access token")}
)
