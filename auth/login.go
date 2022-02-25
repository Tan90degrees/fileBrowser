package auth

import (
	"log"
	"net/http"
)

const (
	LOGIN_SUCCESS      int8 = 0
	LOGIN_NO_USER      int8 = 1
	LOGIN_WRONG_PASSWD int8 = 2
)

func (u *USER_NODE) Login(userName string, passwd []byte) (*http.Cookie, int8, uint8) {
	i := u
	if i.User.UserName == userName {
		if CheckPasswd(passwd, i.User.PassWord) {
			return i.cookie, LOGIN_SUCCESS, i.User.Permission
		} else {
			return nil, LOGIN_WRONG_PASSWD, 0
		}
	}
	for i = i.Next; i != u; i = i.Next {
		if i.User.UserName == userName {
			if CheckPasswd(passwd, i.User.PassWord) {
				return i.cookie, LOGIN_SUCCESS, i.User.Permission
			} else {
				return nil, LOGIN_WRONG_PASSWD, 0
			}
		}
	}
	cookie, db := u.ReadUsers(userName, passwd)
	switch db {
	case USER_OK:
		return cookie, LOGIN_SUCCESS, i.User.Permission
	case NO_USER_FILE:
		log.Println("filebrowser.db was not found")
		return nil, LOGIN_NO_USER, 0
	case BROKEN_USER_FILE:
		log.Println("filebrowser.db was broken")
		return nil, LOGIN_NO_USER, 0
	case NO_USER_MATCH:
		return nil, LOGIN_NO_USER, 0
	case WRONG_PASSWORD:
		return nil, LOGIN_WRONG_PASSWD, 0
	default:
		return nil, LOGIN_NO_USER, 0
	}
}
