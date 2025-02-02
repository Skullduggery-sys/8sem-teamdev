package model

type RoleType int64

const (
	DefaultUser RoleType = iota
	Admin
)

func (s RoleType) String() string {
	switch s {
	case DefaultUser:
		return "user"
	case Admin:
		return "admin"
	}

	return "unknown"
}

type User struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Login       string `json:"login"`
	Role        string `json:"role"`
	Password    string `json:"password"`
	AdminSecret string `json:"admintoken"`
}
