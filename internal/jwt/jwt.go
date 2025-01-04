package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const appSecret = "app_secret"

func NewToken(user string, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user"] = user
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(appSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func CheckToken(token string) (string, error) {
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if ok && tokenParsed.Valid {
		return claims["user"].(string), nil
	}
	return "", errors.New("the token is not valid")
}
