package resolver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal value for digest: %w", err)
	}
	return digestBytes(data), nil
}

func stripDigest(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return data, nil
}
