package hasher

import "fmt"

func ExampleSha256Hasher_Hash() {
	hashGenerator := NewSHA256Hasher()
	hash := hashGenerator.Hash([]byte("test"))
	fmt.Println(hash)
	// Output: [159 134 208 129 136 76 125 101 154 47 234 160 197 90 208 21 163 191 79 27 43 11 130 44 209 93 108 21 176 240 10 8]
}
