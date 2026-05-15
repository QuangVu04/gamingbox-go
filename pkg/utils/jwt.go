package utils

import (
    "errors"
    "time"

    "vault/be/config"

    "github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
    AccessToken  TokenType = "access"
    RefreshToken TokenType = "refresh"
)

type Claims struct {
    UserID   uint     `json:"user_id"`
    Username string   `json:"username"`
    Role     string   `json:"role"`
    Type     TokenType `json:"type"`
    jwt.RegisteredClaims
}

type TokenPair struct {
    AccessToken  string
    RefreshToken string
}

func GenerateTokenPair(userID uint, username string, role string, refreshExpires time.Duration) (*TokenPair, error) {
    accessToken, err := generateToken(userID, username, role, AccessToken, config.App.JWTAccessExpires)
    if err != nil {
        return nil, err
    }

    refreshToken, err := generateToken(userID, username, role, RefreshToken, refreshExpires)
    if err != nil {
        return nil, err
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}

func ParseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, errors.New("unexpected signing method")
            }
            return []byte(config.App.JWTSecret), nil
        },
    )
    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }

    return claims, nil
}

func generateToken(userID uint, username string, role string, tokenType TokenType, expires time.Duration) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        Type:     tokenType,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expires)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "vault/be",
        },
    }

    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.App.JWTSecret))
}