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

//ValidateAddMembers validates add group members request
func ValidateAddMembers(members kube_types.UserGroupMembers) []error {
	var errs []error
	for i, m := range members.Members {
		if m.Access == "" {
			errs = append(errs, fmt.Errorf(isRequiredSlice, "access", i+1))
		}
		if m.Username == "" {
			errs = append(errs, fmt.Errorf(isRequiredSlice, "username", i+1))
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

//ValidateAddMembers validates add group members request
func ValidateUpdateMember(member kube_types.UserGroupMember) []error {
	var errs []error
	if member.Access == "" {
		errs = append(errs, fmt.Errorf(isRequired, "access"))
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
