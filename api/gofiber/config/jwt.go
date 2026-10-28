package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

const (
	jwtPrivatePrefixSuffix = "JWT_PRIVATE_PREFIX_SUFFIX"
	jwtPrivateKey          = "JWT_PRIVATE_KEY"
)

var (
	privateKey *rsa.PrivateKey

	JWTConfig jwtware.Config

	privateKeyPem string
	publicKeyPem  string // temp save in memory, derived from privateKey
)

func NewJWTHandler() fiber.Handler {
	return jwtware.New(JWTConfig)
}

func retrieveJWTPrivateKey() string {
	presuf := strings.SplitN(os.Getenv(jwtPrivatePrefixSuffix), ",", 2)
	if len(presuf) != 2 {
		return ""
	}
	return presuf[0] + "\n" + os.Getenv(jwtPrivateKey) + "\n" + presuf[1]
}

//func RetrieveJWTPublicKey() string {
//	presuf := strings.Split(os.Getenv("JWT_PUBLIC_PREFIX_SUFFIX"), ",")
//	return presuf[0] + "\n" + os.Getenv("JWT_PUBLIC_KEY") + "\n" + presuf[1]
//}

//func RetrievePublicKey() string {
//	return publicKeyPem
//}

func FormJWTInfo() fiber.Map {
	res := fiber.Map{
		"rsa_public_key": publicKeyPem,
	}

	return res
}

// https://stackoverflow.com/questions/13555085/save-and-load-crypto-rsa-privatekey-to-and-from-the-disk
func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)
}

func ExportRsaPublicKeyAsPemStr(privkey *rsa.PrivateKey) (string, error) {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(privkey.Public())
	if err != nil {
		return "", err
	}

	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem), nil
}

func parseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	// Decode the PEM-encoded private key data
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	// Parse the RSA private key
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

// InitializeJWTFromPrivateKeyPEM loads a private key from PEM string, configures JWT middleware,
// and initializes the global JWT state. Returns the parsed private key or an error.
func InitializeJWTFromPrivateKeyPEM(privateKeyPEM string) (*rsa.PrivateKey, error) {
	var err error
	privateKey, err = parseRsaPrivateKeyFromPemStr(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %v", err)
	}

	JWTConfig.SigningKey = jwtware.SigningKey{
		JWTAlg: jwtware.RS512,
		Key:    privateKey.Public(),
	}

	privateKeyPem = ExportRsaPrivateKeyAsPemStr(privateKey)
	publicKeyPem, _ = ExportRsaPublicKeyAsPemStr(privateKey)

	//log.Printf("\n%v\n%v\n", privateKeyPem, publicKeyPem)

	return privateKey, nil
}

// SetupRSAPrivateKey initializes JWT from environment configuration.
// This is a convenience wrapper that retrieves the key from environment variables.
// For better testability, use InitializeJWTFromPrivateKeyPEM directly.
func SetupRSAPrivateKey() (*rsa.PrivateKey, error) {
	privateKeyPEM := retrieveJWTPrivateKey()
	//log.Printf("RetrieveJWTPrivateKey: %+v")
	return InitializeJWTFromPrivateKeyPEM(privateKeyPEM)
}
