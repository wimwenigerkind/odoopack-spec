package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const prefix = "sha256:"

// Digest holds the lowercase hex of a sha256 sum, so two digests of the same
// content compare equal with == whatever case the input used.
type Digest struct {
	hex string
}

func Parse(s string) (Digest, error) {
	algo, h, ok := strings.Cut(s, ":")
	if !ok {
		return Digest{}, fmt.Errorf("digest: %q must be sha256:<64 hex characters>", s)
	}
	if algo != "sha256" {
		return Digest{}, fmt.Errorf("digest: unsupported algorithm %q in %q, only sha256 is supported", algo, s)
	}
	if len(h) != 2*sha256.Size {
		return Digest{}, fmt.Errorf("digest: %q must have %d hex characters, got %d", s, 2*sha256.Size, len(h))
	}
	if _, err := hex.DecodeString(h); err != nil {
		return Digest{}, fmt.Errorf("digest: %q is not hex: %w", s, err)
	}
	return Digest{hex: strings.ToLower(h)}, nil
}

func (d Digest) String() string {
	return prefix + d.hex
}

func (d Digest) Hex() string {
	return d.hex
}

// Verify reads r to the end and reports a mismatch with both sums, so a
// caller can tell a corrupt download from a wrong entry in the index.
func (d Digest) Verify(r io.Reader) error {
	if d.hex == "" {
		return fmt.Errorf("digest: cannot verify against an empty digest")
	}
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return fmt.Errorf("digest: read: %w", err)
	}
	got := Digest{hex: hex.EncodeToString(h.Sum(nil))}
	if got != d {
		return fmt.Errorf("digest: mismatch: expected %s, got %s", d, got)
	}
	return nil
}
