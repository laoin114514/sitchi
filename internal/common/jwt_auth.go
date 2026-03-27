package common

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtAuthService struct {
	SecretKey       string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type TokenClaims struct {
	UserID     int64    `json:"user_id"`
	ModuleCode string   `json:"module_code"`
	Roles      []string `json:"roles,omitempty"`
	TokenType  string   `json:"token_type"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token is expired")
	ErrInvalidConfig    = errors.New("invalid jwt config")
	ErrInvalidTokenType = errors.New("invalid token type")
)

func NewJwtAuthService(secretKey, issuer string, accessTTL, refreshTTL time.Duration) *JwtAuthService {
	return &JwtAuthService{
		SecretKey:       secretKey,
		Issuer:          issuer,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	}
}

func (s *JwtAuthService) GenerateAccessToken(userID int64, moduleCode string, roles []string) (string, error) {
	if err := s.validateConfig(); err != nil {
		return "", err
	}

	now := time.Now()
	claims := TokenClaims{
		UserID:     userID,
		ModuleCode: moduleCode,
		Roles:      roles,
		TokenType:  "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Issuer,
			Subject:   moduleCode,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.AccessTokenTTL)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.SecretKey))
}

func (s *JwtAuthService) GenerateRefreshToken(userID int64, moduleCode string) (string, error) {
	if err := s.validateConfig(); err != nil {
		return "", err
	}

	now := time.Now()
	claims := TokenClaims{
		UserID:     userID,
		ModuleCode: moduleCode,
		TokenType:  "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Issuer,
			Subject:   moduleCode,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.RefreshTokenTTL)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.SecretKey))
}

func (s *JwtAuthService) ParseAndVerifyToken(tokenString string) (*TokenClaims, error) {
	if err := s.validateConfig(); err != nil {
		return nil, err
	}

	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidToken
		}
		return []byte(s.SecretKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Issuer != s.Issuer {
		return nil, ErrInvalidToken
	}
	if claims.TokenType != "access" && claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

func (s *JwtAuthService) ParseAndVerifyAccessToken(tokenString string) (*TokenClaims, error) {
	claims, err := s.ParseAndVerifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "access" {
		return nil, ErrInvalidTokenType
	}
	return claims, nil
}

func (s *JwtAuthService) ParseAndVerifyRefreshToken(tokenString string) (*TokenClaims, error) {
	claims, err := s.ParseAndVerifyToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}
	return claims, nil
}

func (s *JwtAuthService) validateConfig() error {
	if s == nil || strings.TrimSpace(s.SecretKey) == "" || strings.TrimSpace(s.Issuer) == "" || s.AccessTokenTTL <= 0 || s.RefreshTokenTTL <= 0 {
		return ErrInvalidConfig
	}
	return nil
}
