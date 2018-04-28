package validation

import (
	"fmt"

	"git.containerum.net/ch/user-manager/pkg/models"
)

//ValidateUserCreateRequest validates simple send mail request
func ValidateDomain(login models.Domain) []error {
	errs := []error{}
	if login.Domain == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Domain"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
