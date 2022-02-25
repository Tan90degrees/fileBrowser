package auth

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	MOD_SUCCESS     int8  = 0
	MOD_NO_USER     int8  = 1
	FOUND_SUCCESS   int8  = 2
	FOUND_NO_USER   int8  = 3
	COOKIE_GOOD     int8  = 4
	COOKIE_NOTFOUND int8  = 5
	COOKIE_EXPIRED  int8  = 6
	USER_NUM_MAX    uint8 = 128
)

type USER struct {
	UserName string
	// Root       string
	PassWord   []byte
	Permission uint8
}

type USER_NODE struct {
	User   USER
	cookie *http.Cookie
	Pre    *USER_NODE
	Next   *USER_NODE
}

func initCookie(userNode *USER_NODE) {
	userNode.cookie = &http.Cookie{
		Name:    userNode.User.UserName,
		Value:   fmt.Sprintf("%x", md5.Sum([]byte(userNode.User.UserName+time.Now().String()))),
		Path:    "/",
		Expires: time.Now().Add(10 * time.Minute),
		MaxAge:  600,
	}
}

func newUser(userName string, passwd []byte, permission uint8) *USER_NODE {
	user := new(USER_NODE)
	user.User.UserName = userName
	user.User.Permission = permission
	user.User.PassWord = passwd
	initCookie(user)
	return user
}

func InitUserList() *USER_NODE {
	user := new(USER_NODE)
	user.User.UserName = "root"
	user.Pre = user
	user.Next = user
	return user
}

func (u *USER_NODE) AddUser(userName string, passwd []byte, permission uint8) (*http.Cookie, int8) {
	user := newUser(userName, passwd, permission)
	u.Pre.Next = user
	user.Pre = u.Pre
	u.Pre = user
	user.Next = u
	return user.cookie, MOD_SUCCESS
}

func (u *USER_NODE) FoundUser(userName string) (*USER_NODE, int8) {
	i := u
	if i.User.UserName != userName {
		i = u.Next
		for i != u && i.User.UserName != userName {
			i = i.Next
		}
	} else {
		return i, FOUND_SUCCESS
	}
	if i.User.UserName == userName {
		return i, FOUND_SUCCESS
	} else {
		return nil, FOUND_NO_USER
	}
}

func (u *USER_NODE) DelUser(userName string) (*USER_NODE, int8) {
	i := u
	if i.User.UserName != userName {
		i = u.Next
		for i != u && i.User.UserName != userName {
			i = i.Next
		}
	} else if i.Next != i {
		i.Pre.Next = i.Next
		i.Next.Pre = i.Pre
		u = u.Next
		return u, MOD_SUCCESS
	} else {
		return nil, MOD_SUCCESS
	}
	if i.User.UserName == userName {
		i.Pre.Next = i.Next
		i.Next.Pre = i.Pre
		return u, MOD_SUCCESS
	} else {
		return u, MOD_NO_USER
	}
}

func (u *USER_NODE) Check(userName string, cookie string) int8 {
	user, foundFlag := u.FoundUser(userName)
	if foundFlag == FOUND_SUCCESS {
		if time.Now().Before(user.cookie.Expires) {
			if user.cookie.Value == cookie {
				return COOKIE_GOOD
			}
		} else {
			u.DelUser(userName)
			return COOKIE_EXPIRED
		}
	} else {
		return COOKIE_NOTFOUND
	}
	return UNDEFINED_WRONG
}

func (u *USER_NODE) String() string {
	i := u
	var ret []byte
	ret = append(ret, []byte(i.User.UserName+"	"+strconv.FormatInt(int64(i.User.Permission), 10)+"\n")...)
	i = i.Next
	for i != u {
		ret = append(ret, []byte(i.User.UserName+"	"+strconv.FormatInt(int64(i.User.Permission), 10)+"\n")...)
		i = i.Next
	}
	return string(ret)
}
