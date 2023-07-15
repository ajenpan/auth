package common

import (
	"crypto/rsa"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(pk *rsa.PrivateKey, uid int64, uname string) (string, error) {
	claims := make(jwt.MapClaims)
	// TOOD: fix
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	claims["iat"] = time.Now().Unix()
	claims["aud"] = "games"
	claims["nbf"] = time.Now().Unix()
	claims["iss"] = "hotwave-auth"
	claims["sub"] = uname
	claims["uid"] = strconv.FormatInt(uid, 10)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(pk)
}

func VerifyToken(pk *rsa.PublicKey, tokenRaw string) (int64, string, error) {
	claims := make(jwt.MapClaims)
	token, err := jwt.ParseWithClaims(tokenRaw, claims, func(t *jwt.Token) (interface{}, error) {
		return pk, nil
	})
	if err != nil {
		return 0, "", err
	}
	if !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}
	uname := claims["sub"]
	uidstr := claims["uid"]
	uid, _ := strconv.ParseInt(uidstr.(string), 10, 64)
	return uid, uname.(string), nil
}
