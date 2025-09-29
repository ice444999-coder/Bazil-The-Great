
package service

type UserService interface {
	Signup(username, email, password string) error
	Login(username, password string)(id uint,accessToken string, refreshToken string, err error)
	Refresh(refreshToken string) (string, error)
}

