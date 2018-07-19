package validation

import (
	"fmt"

	"git.containerum.net/ch/user-manager/pkg/models"
)

func ValidateLoginRequest(login models.LoginRequest) []error {
	var errs []error
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

func ValidateOAuthLoginRequest(login models.OAuthLoginRequest) []error {
	var errs []error
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
func ValidateLink(link models.Link) []error {
	var errs []error
	if link.Link == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Link"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
