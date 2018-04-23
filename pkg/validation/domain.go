package validation

import (
	"fmt"

	umtypes "git.containerum.net/ch/user-manager/pkg/models"
)

//ValidateUserCreateRequest validates simple send mail request
//nolint: gocyclo
func ValidateDomain(login umtypes.Domain) []error {
	errs := []error{}
	if login.Domain == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Resource"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
