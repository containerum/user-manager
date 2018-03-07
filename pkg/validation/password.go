package validation

import (
	"fmt"

	umtypes "git.containerum.net/ch/json-types/user-manager"
)

func ValidatePasswordChangeRequest(link umtypes.PasswordRequest) []error {
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

func ValidatePasswordRestoreRequest(link umtypes.PasswordRequest) []error {
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
