package jwts

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	UserID   int64 `json:"user_id"`
}