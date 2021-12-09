package crypto

import (
	"crypto/sha512"
)

func HashData(data []byte) (Hash, error) {
	sum := sha512.New()
	if _, err := sum.Write(data); err != nil {
		return nil, err
	}
	return sum.Sum(nil), nil
}
