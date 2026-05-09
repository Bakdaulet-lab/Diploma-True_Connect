// seedhash generates Argon2id password hashes using the same parameters as
// production auth. Use it to produce password_hash values for seed migrations.
//
// Usage:
//
//	go run ./cmd/seedhash/ -password=secret -count=3
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/trueconnect/backend/internal/pkg/crypto"
)

func main() {
	password := flag.String("password", "password", "plaintext password to hash")
	count := flag.Int("count", 1, "number of hashes to generate")
	flag.Parse()

	for i := 0; i < *count; i++ {
		hash, err := crypto.HashPassword(*password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(hash)
	}
}
