package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func Last4(hexStr string) string {
	if len(hexStr) < 4 {
		return hexStr
	}
	return hexStr[len(hexStr)-4:]
}
