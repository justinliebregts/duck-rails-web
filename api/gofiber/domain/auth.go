package domain

var (
	MyHashProvider HashProvider = &BCryptHashPROV{}
)

type HashProvider interface {
	HashPasswordEncoding(eType EncodingType, plain string) (string, error)
}
