package domain

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	myERR "gofiber/errors"

	"golang.org/x/crypto/bcrypt"
)

type HashType string
type EncodingType string

var (
	HashSHA1   HashType = "sha1"
	HashSHA256 HashType = "sha256"

	EncodingNone    EncodingType = ""
	EncodingHex     EncodingType = "hex"
	EncodingBase64  EncodingType = "base64"
	EncodingTypeMap              = map[string]EncodingType{
		string(EncodingNone):   EncodingNone,
		string(EncodingHex):    EncodingHex,
		string(EncodingBase64): EncodingBase64,
	}

	encoderM = map[EncodingType]func([]byte) string{
		EncodingNone:   func(by []byte) string { return string(by) },
		EncodingHex:    BytesToHex,
		EncodingBase64: BytesToBase64,
	}

	decoderM = map[EncodingType]func(string) ([]byte, error){
		EncodingNone:   func(s string) ([]byte, error) { return []byte(s), nil },
		EncodingHex:    HexToBytes,
		EncodingBase64: Base64ToBytes,
	}
)

func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func HexToBytes(h string) ([]byte, error) {
	return hex.DecodeString(h)
}

func BytesToBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Base64ToBytes(h string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(h)
}

// BCryptHashPROV hash string with BCrypt
type BCryptHashPROV struct {
}

func (b *BCryptHashPROV) HashPasswordEncoding(eType EncodingType, plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), 10)
	return encoderM[eType](bytes), err
}

func (b *BCryptHashPROV) ComparePasswordEncoding(eType EncodingType, plain string, hashed string) error {
	decod, _ := decoderM[eType](hashed)
	return bcrypt.CompareHashAndPassword(decod, []byte(plain))
}

// SHA1HashPROV hash string with SHA1
type SHA1HashPROV struct {
}

func (s *SHA1HashPROV) HashPasswordEncoding(eType EncodingType, plain string) (string, error) {
	bytes := computeSHA1([]byte(plain))
	return encoderM[eType](bytes), nil
}

func (s *SHA1HashPROV) ComparePasswordEncoding(eType EncodingType, plain string, hashed string) error {
	h, _ := s.HashPasswordEncoding(eType, plain)
	if subtle.ConstantTimeCompare([]byte(h), []byte(hashed)) == 1 {
		return nil
	}

	return myERR.NoMatch
}

// SHA256HashPROV hash string with SHA256
type SHA256HashPROV struct {
}

func (s *SHA256HashPROV) HashPasswordEncoding(eType EncodingType, plain string) (string, error) {
	bytes := computeSHA256([]byte(plain))
	return encoderM[eType](bytes), nil
}

func (s *SHA256HashPROV) ComparePasswordEncoding(eType EncodingType, plain string, hashed string) error {
	h, _ := s.HashPasswordEncoding(eType, plain)
	if subtle.ConstantTimeCompare([]byte(h), []byte(hashed)) == 1 {
		return nil
	}
	return myERR.NoMatch
}

func computeSHA1(input []byte) []byte {
	hash := sha1.New()
	hash.Write(input)
	return hash.Sum(nil)
}

func computeSHA256(input []byte) []byte {
	hash := sha256.New()
	hash.Write(input)
	return hash.Sum(nil)
}
