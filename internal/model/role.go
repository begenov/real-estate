package model

const (
	Role_Admin   = "admin"
	Role_Manager = "manager"
)

const (
	Role_Admin_ID   = iota + 1
	Role_Manager_ID = 2
)

type Role struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
}

func HasRoles(userRoles []Role, roles ...string) bool {
	roleSet := make(map[string]struct{})
	for i := range userRoles {
		roleSet[userRoles[i].Code] = struct{}{}
	}

	for i := range roles {
		if _, exists := roleSet[roles[i]]; exists {
			return true
		}
	}

	return false
}
