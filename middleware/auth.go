package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"DeNet/utils"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondWithUnauthorized(w, "Authorization header is required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondWithUnauthorized(w, "Invalid authorization header format. Expected: Bearer <token>")
			return
		}

		tokenString := parts[1]
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "your-secret-key-change-in-production"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			log.Printf("JWT parse error: %v", err)
			utils.RespondWithUnauthorized(w, "Invalid or expired token")
			return
		}

		if !token.Valid {
			utils.RespondWithUnauthorized(w, "Invalid or expired token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.RespondWithUnauthorized(w, "Invalid token claims")
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			utils.RespondWithUnauthorized(w, "Invalid token: user_id not found")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


