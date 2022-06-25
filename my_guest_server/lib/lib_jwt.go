package lib

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

const jwt_secret = "lvhpg6HGnYqRyDvn"

var tokenMinute = time.Minute * 30

func GenerateJWT(role string) (string, error) {
	var mySigningKey = []byte(jwt_secret)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["role"] = role
	claims["exp"] = time.Now().Add(tokenMinute).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt.Printf("Something Went Wrong: %s\n", err.Error())
		return "", err
	}
	return tokenString, nil
}

func ParseJWT(token string) error {
	var mySigningKey = []byte(jwt_secret)
	parse, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing")
		}
		return mySigningKey, nil
	})

	if err != nil {
		return err
	}

	if claims, ok := parse.Claims.(jwt.MapClaims); ok && parse.Valid {
		if claims["role"] == "admin" {
			return nil
		} else {
			return errors.New("role is wrong")
		}
	}
	return errors.New("Not Authorized")
}
