package validation

import (
	"fmt"

	umtypes "git.containerum.net/ch/json-types/user-manager"
)

//ValidateLink validates simple send mail request
func ValidateResource(req umtypes.BoundAccountDeleteRequest) []error {
	errs := []error{}
	if req.Resource == "" {
		errs = append(errs, fmt.Errorf(isRequired, "Resource"))
	}
	return nil
}
