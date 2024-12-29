package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var JwtKey = []byte("secret-key")

type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func Auth() gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		cookie, err := ctx.Cookie("session_token")
		if err != nil {
			if strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			} else {
				ctx.Redirect(http.StatusSeeOther, "/login")
			}
			ctx.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(cookie, claims, func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}

		ctx.Set("id", claims.UserID)
		ctx.Next()
	})
}