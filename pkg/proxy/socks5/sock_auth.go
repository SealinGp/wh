package socks5

type SockAuthImpl interface {
	Auth(user, pass string) bool
}

type SockAuth struct {
	user string
	pass string
}

type SockAuthOpt struct {
	User string
	Pass string
}

func NewSockAuth(opt *SockAuthOpt) *SockAuth {
	sockAuth := &SockAuth{
		user: opt.User,
		pass: opt.Pass,
	}
	return sockAuth
}

func (sockAuth *SockAuth) Auth(user, pass string) bool {

	if sockAuth.user != user {
		return false
	}

	return sockAuth.pass == pass
}
