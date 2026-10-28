package v2

import "github.com/go-playground/validator/v10"

var (
	// use a single instance of Validate, it caches struct info
	validate = validator.New()
)
