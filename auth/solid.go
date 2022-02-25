package auth

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
)

const (
	USER_OK          int8 = 0
	NO_USER_FILE     int8 = 1
	BROKEN_USER_FILE int8 = 2
	NO_USER_MATCH    int8 = 3
	WRONG_PASSWORD   int8 = 4
	UNDEFINED_WRONG  int8 = 127
	// MAX_USER_NAME_LEN    uint8  = 64
	// MAX_USER_KEY_LEN     int    = 256
	USER_FILE string = "filebrowser.db"
)

func RegUser() error {
	var userName string
	var passwd []byte
	var permission string
	fmt.Print("Username:")
	fmt.Scanln(&userName)
	fmt.Print("Password:")
	fmt.Scanln(&passwd)
	fmt.Print("Permission:")
	fmt.Scanln(&permission)
	fp, err := os.OpenFile(USER_FILE, os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.WriteString(userName + " " + HashPasswdString(passwd) + " " + permission + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (u *USER_NODE) ReadUsers(userName string, passWord []byte) (*http.Cookie, int8) {
	fp, err := os.Open(USER_FILE)
	if err != nil {
		return nil, NO_USER_FILE
	}
	defer fp.Close()
	fo := bufio.NewReader(fp)
	for {
		name, err := fo.ReadBytes(' ')
		if err != nil {
			return nil, NO_USER_MATCH
		}
		if userName == string(name[:len(name)-1]) {
			key, err := fo.ReadBytes(' ')
			if err != nil {
				return nil, BROKEN_USER_FILE
			}
			if CheckPasswd(passWord, key[:len(key)-1]) {
				var permission uint8
				permission, err = fo.ReadByte()
				if err != nil {
					return nil, BROKEN_USER_FILE
				}
				cookie, _ := u.AddUser(userName, key[:len(key)-1], permission)
				return cookie, USER_OK
			} else {
				fmt.Printf("%s\n%s\n", passWord, key[:len(key)-1])
				return nil, WRONG_PASSWORD
			}
		}
		_, err = fo.ReadBytes('\n')
		if err != nil {
			return nil, BROKEN_USER_FILE
		}
	}
}
