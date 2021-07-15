package key

import "encoding/hex"

func CreateKey() ([]byte, string, error) {
	key, err := GenerateRSAKey()
	if err != nil {
		return nil, "", err
	}

	keyString := hex.EncodeToString(key)

	return key, keyString, nil
}
