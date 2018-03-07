package server

import "git.containerum.net/ch/user-manager/pkg/models"

// CreateFilterFunc is a helper function which creates a function needed to check if profile satisfies given filters
func CreateFilterFunc(filters ...string) func(p models.UserProfileAccounts) bool {
	var filterFuncs []func(p models.UserProfileAccounts) bool
	for _, filter := range filters {
		switch filter {
		case "active":
			filterFuncs = append(filterFuncs, func(p models.UserProfileAccounts) bool {
				return p.User.IsActive
			})
		case "inactive":
			filterFuncs = append(filterFuncs, func(p models.UserProfileAccounts) bool {
				return !p.User.IsActive
			})
		case "in_blacklist":
			filterFuncs = append(filterFuncs, func(p models.UserProfileAccounts) bool {
				return p.User.IsInBlacklist
			})
		case "deleted":
			filterFuncs = append(filterFuncs, func(p models.UserProfileAccounts) bool {
				return p.User.IsDeleted
			})
		case "user":
			filterFuncs = append(filterFuncs, func(p models.UserProfileAccounts) bool {
				return p.User.Role == "user"
			})
		case "admin":
			filterFuncs = append(filterFuncs, func(p models.UserProfileAccounts) bool {
				return p.User.Role == "admin"
			})
		}

	}

	satisfiesFilter := func(p models.UserProfileAccounts) bool {
		ret := true
		for _, v := range filterFuncs {
			ret = ret && v(p)
		}
		return ret
	}

	return satisfiesFilter
}
