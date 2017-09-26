package token

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

//TokenClaims used for JWT
type TokenClaims struct {
	Username string `json:"username,omitempty"`
	ID       int    `json:"user_id,omitempty"`
	jwt.StandardClaims
}

var mySigningKey = []byte("secret")

//getAnonToken will get a token for non-auth user
func getAnonToken() (string, error) {
	claims := TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 5).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	/* Sign the token with secret */
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

//GetTokenHandler will get a token for the username
func getToken(username string, id int) (string, error) {

	claims := TokenClaims{
		username,
		id,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 5).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	/* Sign the token with secret */
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

//ValidateToken will validate the token
func ValidateToken(incomingToken string) (bool, string) {
	token, err := jwt.ParseWithClaims(incomingToken, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(mySigningKey), nil
	})

	if err != nil {
		return false, ""
	}

	claims := token.Claims.(*TokenClaims)
	return token.Valid, claims.Username
}
