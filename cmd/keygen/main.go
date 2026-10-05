package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Utility to generate api random key")
	fmt.Println("---------------------")

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		log.Fatalf("reading error: %v", err)
	}
	rawKey := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(rawKey))

	fmt.Println("Key:")
	fmt.Println(rawKey)
	fmt.Println("Hash:")
	fmt.Println(`\x` + hex.EncodeToString(hash[:]))
}
