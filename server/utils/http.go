package utils

import (
	"testing"

	"github.com/Trojan295/chinchilla-server/server/auth"
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
func SetupRouter(t *testing.T) *gin.Engine {
	r := gin.Default()
	r.Use(auth.JWTToken("secret"))
	return r
}
