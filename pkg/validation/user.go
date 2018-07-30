package validation

import (
	"fmt"

	"git.containerum.net/ch/user-manager/pkg/models"
	"github.com/goware/emailx"
)

//ValidateUserCreateRequest validates simple send mail request
func ValidateUserCreateRequest(user models.RegisterRequest) []error {
	var errs []error
	if user.Login == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Login"))
	} else {
		if err := emailx.ValidateFast(user.Login); err != nil {
			errs = append(errs, err)
		}
	}
	if user.Password == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Password"))
	}
	if user.ReCaptcha == "" {
		errs = append(errs, fmt.Errorf(isRequired, "ReCaptcha"))
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ValidateUserID(user models.UserLogin) []error {
	var errs []error
	if user.ID == "" {
		errs = append(errs, fmt.Errorf(isRequired, "ID"))
	} else {
		if !IsValidUUID(user.ID) {
			errs = append(errs, errInvalidID)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ValidateUserLogin(user models.UserLogin) []error {
	var errs []error
	if user.Login == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Login"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ValidateUserData(user map[string]interface{}) []error {
	errs := []error{}
	if user["email"] != nil {
		if err := emailx.ValidateFast(user["email"].(string)); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
