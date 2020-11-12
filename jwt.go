package helpers

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"os"
)

type Jwt struct{}

func (j Jwt) VerifyToken(tokenString string, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if secret == "" {
			secret = os.Getenv("ACCESS_SECRET")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (j *Jwt) ParseUnverified(tokenString string) (jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok != true {
		return nil, errors.New("Ошибка получения данных из jwt токена")
	}
	return claims, nil
}
