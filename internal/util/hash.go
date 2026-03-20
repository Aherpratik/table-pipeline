package util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func HashRow(row map[string]string) (string, error) {
	b, err := json.Marshal(row)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
