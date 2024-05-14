// Package key package provides a type to generate and print (for example in
// hex) a random SHA256 key
package key

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/eknkc/basex"
)

// ErrorKeyAlreadyGenerated Error occures when trying to generate a key on a
// Key object which already has a generated key
var ErrorKeyAlreadyGenerated = errors.New("key already generated")

// ErrorKeyGenerateFailed Error occures when the key generation failed
var ErrorKeyGenerateFailed = errors.New("Key generation failed")

// ErrorInvalidKey Error occures when the key is invalid
var ErrorInvalidKey = errors.New("invalid key")

// SizeAES256 the byte size required for aes 256 encoding
const SizeAES256 int = 32

type encoding int

const (
	HexEncoding encoding = iota
	Base62Encoding
)

var base62Encoder *basex.Encoding
var encodingType encoding

func init() {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base62Encoder, _ = basex.NewEncoding(alphabet)
}

// NewKey creates a Key object
func NewKey() *Key {
	var k Key
	return &k
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

// Set sets the key
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

// String returns the key as a string
func (k *Key) String() string {
	if encodingType == Base62Encoding {
		return k.toBase62()
	}

	return k.toHex()
}

// FromString creates a Key from a string
// If the string is 64 characters long, it is assumed to be a hex string
// If the string is 43 characters long, it is assumed to be a base62 string
// Otherwise, it returns an error
func FromString(s string) (*Key, error) {
	if len(s) == 64 {
		return FromHex(s)
	}
	if len(s) == 43 {
		return FromBase62(s)
	}
	return nil, ErrorInvalidKey
}

func (k *Key) toBase62() string {
	return base62Encoder.Encode(*k)
}

// FromBase62 creates a Key from a base62 string
func FromBase62(s string) (*Key, error) {
	decoded, err := base62Encoder.Decode(s)

	if err != nil {
		return nil, err
	}

	k := NewKey()
	if err := k.Set(decoded); err != nil {
		return nil, err
	}

	return k, nil
}

// FromHex creates a Key from a hex string
func FromHex(s string) (*Key, error) {
	decoded, err := hex.DecodeString(s)

	if err != nil {
		return nil, err
	}

	k := NewKey()
	if err := k.Set(decoded); err != nil {
		return nil, err
	}

	return k, nil
}

// SetEncodingType sets the encoding type
func SetEncodingType(encoding encoding) {
	encodingType = encoding
}
