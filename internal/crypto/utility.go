package crypto

import (
	"crypto/sha256"
)

func HashData(data []byte) (Hash, error) {
	sum := sha256.New()
	if _, err := sum.Write(data); err != nil {
		return nil, err
	}
	return sum.Sum(nil), nil
}
