package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"log_collect/config"
	"log_collect/internal/infrastructure/hasher"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("%v", err)
	}

	bcryptHasher := hasher.NewBcryptHasher(cfg.Hasher.Cost)

	fmt.Println("Utility to get bcrypt hash from plain text")
	fmt.Println("---------------------")
	fmt.Print("Enter plain text -> ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		log.Fatalf("no input")
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading input: %v", err)
	}

	hash, err := bcryptHasher.Hash(scanner.Text())
	if err != nil {
		log.Fatalf("error hashing: %v", err)
	}

	fmt.Println(hash)
}
