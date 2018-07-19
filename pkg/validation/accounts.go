package validation

import (
	"fmt"

	"git.containerum.net/ch/user-manager/pkg/models"
)

//ValidateLink validates simple send mail request
func ValidateResource(req models.BoundAccountDeleteRequest) []error {
	var errs []error
	if req.Resource == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Resource"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
