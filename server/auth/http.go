package auth

import (
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWTToken is a gin middleware to validate JWT tokens
func JWTToken(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return
		}

		tokenString := strings.Split(authHeader, " ")[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil {
			return
		}

		c.Set("permissions", token.Claims.(jwt.MapClaims)["permissions"])
		c.Set("userID", token.Claims.(jwt.MapClaims)["sub"])
	}
}

// LoginRequired is a gin middleware, which ensures user is logged in
func LoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("userID"); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
			c.Abort()
		}
	}
}

// Auth0Permission is a gin authorization middleware
// for validating Auth0 permissions
func Auth0Permission(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissionsObj, ok := c.Get("permissions")
		permissions := permissionsObj.([]interface{})

		if ok == false {
			c.JSON(http.StatusForbidden, gin.H{"error": "Missing scopes"})
			c.Abort()
		}

		contains := false
		for _, s := range permissions {
			if scope == s.(string) {
				contains = true
			}
		}

		if !contains {
			c.JSON(http.StatusForbidden, gin.H{"error": "Missing scopes"})
			c.Abort()
		}

	}
}
