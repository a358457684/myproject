package util

import (
	"crypto/md5"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func Makepwd(pwd string) string {
	password, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost) //bcrytp加密
	return string(password)
}

func Makemd5(pwd string) string {
	data := []byte(pwd)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func Checkpwd(old string, pwd string) bool {
	//new, _ := bcrypt.GenerateFromPassword([]byte(old), bcrypt.DefaultCost) //bcrytp加密
	//if string(new) == pwd {
	//	return true
	//}

	if err := bcrypt.CompareHashAndPassword([]byte(pwd), []byte(old)); err != nil {
		return false
	}
	return true
}
