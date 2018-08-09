package server

import (
	"git.containerum.net/ch/user-manager/pkg/db"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
)

// CreateFilterFunc is a helper function which creates a function needed to check if profile satisfies given filters
func CreateFilterFunc(filters ...string) func(p db.UserProfileAccounts) bool {
	var filterFuncs []func(p db.UserProfileAccounts) bool
	for _, filter := range filters {
		switch filter {
		case "active":
			filterFuncs = append(filterFuncs, func(p db.UserProfileAccounts) bool {
				return p.User.IsActive
			})
		case "inactive":
			filterFuncs = append(filterFuncs, func(p db.UserProfileAccounts) bool {
				return !p.User.IsActive
			})
		case "in_blacklist":
			filterFuncs = append(filterFuncs, func(p db.UserProfileAccounts) bool {
				return p.User.IsInBlacklist
			})
		case "deleted":
			filterFuncs = append(filterFuncs, func(p db.UserProfileAccounts) bool {
				return p.User.IsDeleted
			})
		case "user":
			filterFuncs = append(filterFuncs, func(p db.UserProfileAccounts) bool {
				return p.User.Role == m.RoleUser
			})
		case "admin":
			filterFuncs = append(filterFuncs, func(p db.UserProfileAccounts) bool {
				return p.User.Role == m.RoleAdmin
			})
		}

	}

	satisfiesFilter := func(p db.UserProfileAccounts) bool {
		ret := true
		for _, v := range filterFuncs {
			ret = ret && v(p)
		}
		return ret
	}

	return satisfiesFilter
}
