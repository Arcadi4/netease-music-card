package netease

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
)

var (
	iv        = []byte("0102030405060708")
	presetKey = []byte("0CoJUm6Qyw8W8jud")
	stdChars  = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	publicKey = []byte("-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDgtQn2JZ34ZC28NWYpAUd98iZ37BUrX/aKzmFbt7clFSs6sXqHauqKWqdtLkF2KexO40H1YTX8z2lSgBBOAxLsvaklV8k4cBFK9snQXE9/DDaFt6Rr7iVZMldczhC0JNgTz+SHXT6CBHuX3e9SdB1Ua44oncaTWz7OBGLbCiK45wIDAQAB\n-----END PUBLIC KEY-----")
)

// EncryptWeapi encrypts request parameters using weapi protocol (AES-128-CBC + RSA)
func EncryptWeapi(data map[string]interface{}) (map[string]string, error) {
	text, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	secretKey, reversedKey := generateRandomKey()

	// First encryption with preset key
	firstEncrypted, err := aesCBCEncrypt(text, presetKey, iv)
	if err != nil {
		return nil, fmt.Errorf("first encryption: %w", err)
	}
	firstBase64 := base64.StdEncoding.EncodeToString(firstEncrypted)

	// Second encryption with random key
	secondEncrypted, err := aesCBCEncrypt([]byte(firstBase64), reversedKey, iv)
	if err != nil {
		return nil, fmt.Errorf("second encryption: %w", err)
	}

	// RSA encrypt the secret key
	encSecKey, err := rsaEncrypt(secretKey, publicKey)
	if err != nil {
		return nil, fmt.Errorf("rsa encryption: %w", err)
	}

	return map[string]string{
		"params":    base64.StdEncoding.EncodeToString(secondEncrypted),
		"encSecKey": hex.EncodeToString(encSecKey),
	}, nil
}

func generateRandomKey() ([]byte, []byte) {
	key := make([]byte, 16)
	reversed := make([]byte, 16)
	for i := 0; i < 16; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(62))
		key[i] = stdChars[n.Int64()]
		reversed[15-i] = stdChars[n.Int64()]
	}
	return key, reversed
}

func aesCBCEncrypt(plaintext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext = pkcs7Pad(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

func rsaEncrypt(data, pubKey []byte) ([]byte, error) {
	// Pad to 128 bytes (prepend zeros)
	padded := make([]byte, 128)
	copy(padded[128-len(data):], data)

	block, _ := pem.Decode(pubKey)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub := pubInterface.(*rsa.PublicKey)

	// Manual RSA encryption (no padding, as per Netease protocol)
	c := new(big.Int).SetBytes(padded)
	encrypted := c.Exp(c, big.NewInt(int64(pub.E)), pub.N).Bytes()

	return encrypted, nil
}
