package helpers

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"os"
)

type Jwt struct{}

func (j Jwt) VerifyTokenHMAC(tokenString string, secret string) (*jwt.Token, error) {
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

func (j Jwt) VerifyTokenRSA(tokenString string, publicKey string) (*jwt.Token, error) {
	if publicKey == "" {
		return nil, errors.New("public key must not be empty")
	}
	rsaPublicKey, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return rsaPublicKey, nil
		}
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
