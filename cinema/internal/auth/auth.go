package auth

import "golang.org/x/crypto/bcrypt"

//func Hash(password string) string {
//	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//	return string(hashed)
//}

func CheckPassword(plain, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}
