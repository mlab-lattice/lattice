package users

type User interface {
	GetName() string
}

type DefaultUser struct {
	Username string
}

func (user *DefaultUser) GetName() string {
	return user.Username
}
