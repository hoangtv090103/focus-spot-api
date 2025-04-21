package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
    ErrInvalidToken = errors.New("token is invalid")
    ErrExpiredToken = errors.New("token has expired")
)

// Claims is the payload in JWT token
type Claims struct {
    UserID string `json:"user_id"`
    Email string `json:"email"`
    jwt.RegisteredClaims
}

type Maker interface{
    // CreateToken creates a new token for userId and email with expire time
    CreateToken(userID, email string, duration time.Duration) (string, error)
    
    // VerifyToken verifies token and return claims
    VerifyToken(token string) (*Claims, error)
}

type JWTMaker struct {
    secretKey string
}

// NewJWTMaker creates a new JWTMaker with secretKey
func NewJWTMaker(secretKey string) (Maker, error) {
    if len(secretKey) < 32 {
        return nil, errors.New("secret key must be at least 32 characters")
    }
    return &JWTMaker{
        secretKey: secretKey,
    }, nil
}

func (maker *JWTMaker) CreateToken(userID, email string, duration time.Duration) (string, error) {
    claims := &Claims {
        UserID: userID,
        Email: email,
        RegisteredClaims: jwt.RegisteredClaims {
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
            IssuedAt: jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(maker.secretKey))
}

func (maker *JWTMaker) VerifyToken(tokenString string) (*Claims, error) {
    keyFunc := func(token *jwt.Token) (interface{}, error) {
        
        // New syntax
        _, ok := token.Method.(*jwt.SigningMethodHMAC)
        if !ok {
            return nil, ErrInvalidToken
        }
        
        return []byte(maker.secretKey), nil
    }
    
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc)
    if err != nil {
        verr, ok := err.(*jwt.ValidationError)
        if ok && errors.Is(verr.Inner, ErrExpiredToken) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok {
        return nil, ErrInvalidToken
    }
    return claims, nil 
}