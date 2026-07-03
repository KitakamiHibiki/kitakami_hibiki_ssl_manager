package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	DeployKey string `json:"deploy_key"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, username, role, email, secret, deployKey string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		Email:     email,
		DeployKey: deployKey,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenStr, secret, deployKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.DeployKey != deployKey {
		return nil, errors.New("deploy key mismatch")
	}
	return claims, nil
}
