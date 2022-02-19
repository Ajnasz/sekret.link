package encrypter

// Encrypter interface to encrypt and decrypt byte arrays
type Encrypter interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}
