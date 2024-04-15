package key

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

// ErrorKeyAlreadyGenerated Error occures when trying to generate a key on a
// Key object which already has a generated key
var ErrorKeyAlreadyGenerated = errors.New("key already generated")
var ErrorKeyGenerateFailed = errors.New("Key generation failed")
var ErrorInvalidKey = errors.New("invalid key")

// SizeAES256 the byte size required for aes 256 encoding
const SizeAES256 int = 32

// NewKey creates a Key struct
func NewKey() *Key {
	return &Key{}
}

// NewGeneratedKey Creates a Key and fills the key
func NewGeneratedKey() (*Key, error) {
	k := NewKey()
	if err := k.Generate(); err != nil {
		return nil, errors.Join(ErrorKeyGenerateFailed, err)
	}

	return k, nil
}

// Key is type to generate and print (for example n hex) a random key
type Key []byte

func (k Key) generateRandomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}

// Generate Generates creates the key, returns error if the key generation
// failed
func (k *Key) Generate() error {
	if len(*k) != 0 {
		return ErrorKeyAlreadyGenerated
	}
	key, err := k.generateRandomBytes(SizeAES256)
	if err != nil {
		return err
	}

	*k = key
	return nil
}

// Get returns the key
func (k *Key) Get() []byte {
	return *k
}

func (k *Key) Set(key []byte) error {
	if len(key) != SizeAES256 {
		return ErrorInvalidKey
	}

	*k = key

	return nil
}

// toHex Converts the key to hex string
func (k *Key) toHex() string {
	return hex.EncodeToString(*k)
}

func (k *Key) String() string {
	return k.toHex()
}

func FromHex(s string) (*Key, error) {
	byte, err := hex.DecodeString(s)

	if err != nil {
		return nil, err
	}

	k := NewKey()
	if err := k.Set(byte); err != nil {
		return nil, err
	}

	return k, nil
}
