package utils

import (
	"github.com/Trojan295/chinchilla/server/auth"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// BuildToken func
func BuildToken(claims map[string]interface{}) string {
	jwtClaims := jwt.MapClaims(claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, _ := token.SignedString([]byte("secret"))
	return tokenString
}

// SetupRouter func
func SetupRouter() *gin.Engine {
	r := gin.Default()
	auth.SetupAuthentication(r, map[string]interface{}{
		"type": "jwt",
		"key":  "secret",
	})
	return r
}
