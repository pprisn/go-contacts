package models

import (
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	u "github.com/pprisn/go-contacts/utils"
	"golang.org/x/crypto/bcrypt"
)

/*
JWT claims struct
*/
type Token struct {
	UserId uint
	jwt.StandardClaims
}

//a struct to rep user account
type Account struct {
	gorm.Model
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token";sql:"-"`
	CodValid string `json:"codvalid"`
}

//Validate incoming user details...
func (account *Account) Validate() (map[string]interface{}, bool) {

	if !strings.Contains(account.Email, "@") {
		return u.Message(false, "Email address is required"), false
	}

	if len(account.Password) < 6 {
		return u.Message(false, "Password is required"), false
	}

	//Email must be unique
	temp := &Account{}

	//check for errors and duplicate emails
	err := GetDB().Table("accounts").Where("email = ?", account.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error. Please retry"), false
	}
	if temp.Email != "" {
		return u.Message(false, "Email address already in use by another user."), false
	}

	return u.Message(false, "Requirement passed"), true
}

func (account *Account) Create() map[string]interface{} {

	if resp, ok := account.Validate(); !ok {
		return resp
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	account.CodValid = u.GenCodeValid(6) //Creating a new code value
	u.SendSmtp(account.Email, "Temporary confirmation code from API", "This is a confirmation code from API.\n"+account.CodValid)
	GetDB().Create(account)

	if account.ID <= 0 {
		return u.Message(false, "Failed to create account, connection error.")
	}

	//Create new JWT token for the newly registered account
	//tk := &Token{UserId: account.ID}
	//Добавим временное ограничение действия токена
	//tk := &Token{
	//		UserId: account.ID,
	//		StandardClaims: jwt.StandardClaims{
	//			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(), Issuer: "test",
	//		},
	//	}

	//	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	//	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	//	account.Token = tokenString
	account.Token = ""

	account.Password = "" //delete password

	response := u.Message(true, "Account has been created, Please confirm registration by sending a verification code at the next login.")
	response["account"] = account
	return response
}

func Login(email, password string, codevalid string) map[string]interface{} {

	account := &Account{}
	err := GetDB().Table("accounts").Where("email = ?", email).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "Email address not found")
		}
		return u.Message(false, "Connection error. Please retry")
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
		return u.Message(false, "Invalid login credentials. Please try again")
	}

	//Worked! Logged In
	account.Password = ""

	if account.CodValid != codevalid {
		account.CodValid = u.GenCodeValid(6) //Creating a new code value
		GetDB().Model(&account).Where("email = ?", email).Update("codvalid", account.CodValid)
		GetDB().Model(&account).Commit()
		u.SendSmtp(account.Email, "Temporary confirmation code from API", "This is a confirmation code from API.\nUse it the next time you log in\n"+account.CodValid)
		return u.Message(false, "A verification code has been sent to you email, use it the next time you log in.")
	}

	//Worked! Logged In
	account.CodValid = ""

	//Create JWT token
	//tk := &Token{UserId: account.ID}
	//Добавим временное ограничение действия токена
	tk := &Token{
		UserId: account.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(), Issuer: "test",
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString //Store the token in the response

	resp := u.Message(true, "Logged In")
	resp["account"] = account
	return resp
}

func GetUser(u uint) *Account {
	acc := &Account{}
	GetDB().Table("accounts").Where("id = ?", u).First(acc)
	if acc.Email == "" { //User not found!
		return nil
	}

	acc.Password = ""
	return acc
}

func GetUsers() []*Account {
	accs := make([]*Account, 0)
	err := GetDB().Table("accounts").Find(&accs).Error
	if err != nil { //Accounts not found!
		return nil
	}
	return accs
}
