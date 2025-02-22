package jwtService

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("SECRET_KEY"))

type UserToken struct {
	UserId   string
	Name     string
	Username string
	Avatar   *string
}

func DecodeToken(tokenString string) (*UserToken, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao analisar o token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := &UserToken{
			UserId:   claims["user_id"].(string),
			Name:     claims["name"].(string),
			Username: claims["username"].(string),
		}
		return user, nil
	}

	return nil, fmt.Errorf("não foi possível extrair dados do token")
}

func CreateToken(user *UserToken) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id":  user.UserId,
			"name":     user.Name,
			"username": user.Username,
			"exp":      time.Now().Add((time.Hour * 48)).Unix(), //two days
		})

	tokeString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokeString, err
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
