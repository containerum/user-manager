package validation

import (
	"fmt"

	kube_types "github.com/containerum/kube-client/pkg/model"
)

//ValidateCreateGroup validates create group request
func ValidateCreateGroup(group kube_types.UserGroup) []error {
	var errs []error
	if group.Label == "" {
		errs = append(errs, fmt.Errorf(isRequired, "label"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
