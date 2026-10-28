package model

import (
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
)

func (ses *Session) FormJWTClaims(privateKey *rsa.PrivateKey) (string, error) {

	// Create the Claims
	claims := jwt.MapClaims{
		"session_id":               -1,
		"gr_id":                    ses.ComposedID(),
		"access_token":             ses.AccessToken,
		"access_token_expires_in":  AccessTokenExpiry.Seconds(),
		"refresh_token":            ses.RefreshToken,
		"refresh_token_expires_in": RefreshTokenExpiry.Seconds(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	// Generate encoded token and send it as response.
	return token.SignedString(privateKey)
}
