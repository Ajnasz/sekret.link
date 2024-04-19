package key

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKey(t *testing.T) {
	k := NewKey()

	err := k.Generate()
	if err != nil {
		t.Fatal("Generate returned error on first call")
	}
	err = k.Generate()

	if err != ErrorKeyAlreadyGenerated {
		t.Fatalf("Generate returned error %s, expected %s", err, ErrorKeyAlreadyGenerated)
	}

	bytesKey := k.Get()

	if len(bytesKey) != 32 {
		t.Fatalf("Expected k.Get() return a 32 length byte slice")
	}

	hexStr := k.toHex()
	isHex, err := regexp.MatchString(`^[0-9a-f]{64}$`, hexStr)

	if err != nil {
		t.Error(err)
	}

	if !isHex {
		t.Fatalf("expected %s to match hex string regexp", hexStr)
	}

	str := k.String()
	assert.Equal(t, hexStr, str, "Stringer interface expected to return hex value")

	// base62 := k.toBase62()

	// isBase62, err := regexp.MatchString(`^[0-9a-zA-Z]{43}$`, base62)

	// assert.NoError(t, err)
	// if !isBase62 {
	// 	t.Fatalf("expected %s to match base62 string regexp", base62)
	// }
}
