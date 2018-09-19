package user

type User interface {
	GetName() string
}

type DefaultUser struct {
	Username string
}

func NewDefaultUser(name string) User {
	return &DefaultUser{
		Username: name,
	}
}

func (user *DefaultUser) GetName() string {
	return user.Username
}
