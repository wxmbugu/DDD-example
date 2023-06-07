package services

const HeaderXPermissions = "X-permissions"

type Checker interface {
	IsSatisfied([]string) bool
}

// And requires all persmissions to match
type And struct {
	Permissions []string
}

// isSatisfied checks if all of the required permissions have been satisfied
func (a And) IsSatisfied(xPerms []string) bool {
	if len(a.Permissions) == 0 {
		return false
	}

	list := make(map[string]struct{}, len(xPerms))
	for _, perm := range xPerms {
		list[perm] = struct{}{}
	}
	for _, perm := range a.Permissions {
		if _, ok := list[perm]; !ok {
			return false
		}
	}
	return true
}

// Or requires at least one permission match.
type Or struct {
	Permissions []string
}

// isSatisfied checks if at least one of the required permissions have been satisfied
func (o Or) IsSatisfied(xPerms []string) bool {
	if len(o.Permissions) == 0 {
		return false
	}
	list := make(map[string]struct{}, len(xPerms))
	for _, perm := range xPerms {
		list[perm] = struct{}{}
	}
	for _, perm := range o.Permissions {
		if _, ok := list[perm]; ok {
			return true
		}
	}
	return false
}
