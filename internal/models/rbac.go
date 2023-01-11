package models

type (
	Roles struct {
		Roleid int
		Role   string
	}
	Permissions struct {
		Permissionid int
		Permission   string
		Roleid       int
	}
	Users struct {
		Id       int
		Email    string
		Password string
		Roleid   int
	}
)
