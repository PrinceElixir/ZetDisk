package utils

import (
    "time"
    "github.com/golang-jwt/jwt/v4"
    "filestorage-backend/config"
)

type Claims struct {
    Email string `json:"email"`
    jwt.RegisteredClaims
}

func GenerateJWT(email string) (string, error) {
    secret := config.Get("JWT_SECRET")
    claims := &Claims{
        Email: email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenStr string) (*Claims, error) {
    secret := config.Get("JWT_SECRET")
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    if err != nil {
        return nil, err
    }
    if !token.Valid {
        return nil, err
    }
    return claims, nil
}


