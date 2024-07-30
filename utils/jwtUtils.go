package utils

import (
    "github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
    UserID string `json:"user_id"`
    jwt.StandardClaims
}

var JWTSecret = []byte("kCMsWstIhJRGsnRSo1FkqSvZrbgt_WMOnb4i5Gb9AXo-")
