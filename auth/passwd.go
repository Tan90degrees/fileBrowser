package auth

import "golang.org/x/crypto/bcrypt"

func HashPasswdString(passwd []byte) string {
	ret, err := bcrypt.GenerateFromPassword(passwd, bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(ret)
}

func HashPasswdBytes(passwd []byte) []byte {
	ret, err := bcrypt.GenerateFromPassword(passwd, bcrypt.DefaultCost)
	if err != nil {
		return nil
	}
	return ret
}

func CheckPasswd(inputPasswd, solidPasswdstring []byte) bool {
	err := bcrypt.CompareHashAndPassword(solidPasswdstring, inputPasswd)
	return err == nil
}
