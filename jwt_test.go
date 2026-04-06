package kstypes

import (
	"crypto/ed25519"
	"os"
	"testing"
)

func TestLoadEd25519Keys(t *testing.T) {
	priv, err := LoadEd25519PrivateKey("testdata/test_private.pem")
	if err != nil {
		t.Fatalf("load private key: %v", err)
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Fatalf("private key size: got %d, want %d", len(priv), ed25519.PrivateKeySize)
	}

	pub, err := LoadEd25519PublicKey("testdata/test_public.pem")
	if err != nil {
		t.Fatalf("load public key: %v", err)
	}
	if len(pub) != ed25519.PublicKeySize {
		t.Fatalf("public key size: got %d, want %d", len(pub), ed25519.PublicKeySize)
	}
}

func TestLoadEd25519PrivateKey_NotFound(t *testing.T) {
	_, err := LoadEd25519PrivateKey("testdata/nonexistent.pem")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadEd25519PublicKey_InvalidPEM(t *testing.T) {
	bad := t.TempDir() + "/bad.pem"
	_ = os.WriteFile(bad, []byte("not a pem"), 0644)
	_, err := LoadEd25519PublicKey(bad)
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}
