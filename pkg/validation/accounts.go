package validation

import (
	"fmt"

	umtypes "git.containerum.net/ch/user-manager/pkg/models"
)

//ValidateLink validates simple send mail request
func ValidateResource(req umtypes.BoundAccountDeleteRequest) []error {
	errs := []error{}
	if req.Resource == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Resource"))
	}
	return nil
}
