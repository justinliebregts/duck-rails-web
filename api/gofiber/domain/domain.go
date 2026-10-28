package domain

import (
	"github.com/go-playground/validator/v10"
)

var (
	// use a single instance of Validate, it caches struct info
	validate = validator.New()

	//LocalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func GetValidator() *validator.Validate {
	return validate
}
