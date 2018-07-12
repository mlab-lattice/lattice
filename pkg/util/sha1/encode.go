package sha1

import (
	"crypto/sha1"
	"encoding/hex"
)

func EncodeToHexString(data []byte) (string, error) {
	h := sha1.New()
	if _, err := h.Write(data); err != nil {
		return "", err
	}

	definitionHash := hex.EncodeToString(h.Sum(nil))
	return definitionHash, nil
}
