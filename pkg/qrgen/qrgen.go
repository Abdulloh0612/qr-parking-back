package qrgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

// GenerateCode returns a random 24-character hex string used as a QR code identifier.
func GenerateCode() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GenerateQRImageURL encodes the given URL into a PNG QR code of the specified size.
func GenerateQRImageURL(url string, size int) ([]byte, error) {
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("encode qr: %w", err)
	}
	return png, nil
}
