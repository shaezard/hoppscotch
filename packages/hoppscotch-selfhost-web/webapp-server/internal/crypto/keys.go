package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

type KeyPair struct {
	SigningKey   ed25519.PrivateKey
	VerifyingKey ed25519.PublicKey
}

func GenerateKeyPair() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	kp := &KeyPair{
		SigningKey:   priv,
		VerifyingKey: pub,
	}

	log.Printf("Generated new signing key pair")
	
	if os.Getenv("GO_ENV") == "development" {
		log.Printf("WARNING: Development mode - Signing key: %s", base64.StdEncoding.EncodeToString(priv))
	}
	
	log.Printf("Verifying key: %s", base64.StdEncoding.EncodeToString(pub))

	return kp, nil
}
