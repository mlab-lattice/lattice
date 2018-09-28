package user

type User interface {
	Name() string
}

type DefaultUser struct {
	name string
}

func NewDefaultUser(name string) User {
	return &DefaultUser{
		name: name,
	}
}

func (u *DefaultUser) Name() string {
	return u.name
}
