package validation

import (
	"fmt"

	umtypes "git.containerum.net/ch/json-types/user-manager"
)

func ValidateLoginRequest(login umtypes.LoginRequest) []error {
	errs := []error{}
	if login.Login == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Login"))
	}
	if login.Password == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Password"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ValidateOAuthLoginRequest(login umtypes.OAuthLoginRequest) []error {
	errs := []error{}
	if login.Resource == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Resource"))
	}
	if login.AccessToken == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Access token"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

//ValidateLink validates simple send mail request
func ValidateLink(link umtypes.Link) []error {
	errs := []error{}
	if link.Link == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Link"))
	}
	return nil
}
