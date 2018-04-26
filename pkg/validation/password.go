package validation

import (
	"fmt"

	"git.containerum.net/ch/user-manager/pkg/models"
)

func ValidatePasswordChangeRequest(link models.PasswordRequest) []error {
	errs := []error{}
	if link.CurrentPassword == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Current password"))
	}
	if link.NewPassword == "" {
		errs = append(errs, fmt.Errorf(isRequired, "New password"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func ValidatePasswordRestoreRequest(link models.PasswordRequest) []error {
	errs := []error{}
	if link.Link == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Link"))
	}
	if link.NewPassword == "" {
		errs = append(errs, fmt.Errorf(isRequired, "New password"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
