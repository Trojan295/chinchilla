package auth

import (
	"log"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// SetupAuthentication registers the propert authentication
// mechanism based on the Configuration
func SetupAuthentication(router *gin.Engine, authConfig map[string]interface{}) {
	if authConfig["type"] == "jwt" {
		log.Println("Using JWT based authentication")

		secret := authConfig["key"].(string)
		router.Use(jwtToken(secret))
	} else if authConfig["type"] == "header" {
		log.Println("Using header based authentication")

		router.Use(headerAuth)
	} else {
		panic("Wrong authentication config")
	}
}

func headerAuth(c *gin.Context) {
	userHeader := c.GetHeader("x-user")
	if userHeader != "" {
		c.Set("userID", userHeader)
	}

	permissionsHeader := c.GetHeader("x-permissions")
	if permissionsHeader != "" {
		permissions := strings.Split(permissionsHeader, ",")
		c.Set("permissions", permissions)
	}

}

// JWTToken is a gin middleware to validate JWT tokens
func jwtToken(secret string) gin.HandlerFunc {
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

		c.Set("userID", token.Claims.(jwt.MapClaims)["sub"])

		if permissionsObj, ok := token.Claims.(jwt.MapClaims)["permissions"].([]interface{}); ok {
			permissions := make([]string, 0)
			for _, permission := range permissionsObj {
				permissions = append(permissions, permission.(string))
			}

			c.Set("permissions", permissions)
		}
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
		permissions := permissionsObj.([]string)

		if ok == false {
			c.JSON(http.StatusForbidden, gin.H{"error": "Missing scopes"})
			c.Abort()
		}

		contains := false
		for _, s := range permissions {
			if scope == s {
				contains = true
			}
		}

		if !contains {
			c.JSON(http.StatusForbidden, gin.H{"error": "Missing scopes"})
			c.Abort()
		}

	}
}
