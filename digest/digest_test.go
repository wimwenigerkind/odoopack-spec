package digest_test

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/wimwenigerkind/odoopack-spec/digest"
)

// sha256 of "hello".
const helloHex = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

func TestParse(t *testing.T) {
	valid := []struct {
		in   string
		want string
	}{
		{"sha256:" + helloHex, "sha256:" + helloHex},
		{"sha256:" + strings.ToUpper(helloHex), "sha256:" + helloHex},
		{"sha256:" + strings.Repeat("0", 64), "sha256:" + strings.Repeat("0", 64)},
	}
	for _, tc := range valid {
		d, err := digest.Parse(tc.in)
		if err != nil {
			t.Errorf("Parse(%q) = %v, want nil", tc.in, err)
			continue
		}
		if d.String() != tc.want {
			t.Errorf("Parse(%q).String() = %q, want %q", tc.in, d.String(), tc.want)
		}
		if d.Hex() != strings.TrimPrefix(tc.want, "sha256:") {
			t.Errorf("Parse(%q).Hex() = %q, want %q", tc.in, d.Hex(), strings.TrimPrefix(tc.want, "sha256:"))
		}
	}

	invalid := []string{
		"",
		helloHex,
		"sha256:",
		"sha256:" + helloHex[:63],
		"sha256:" + helloHex + "0",
		"sha256:" + helloHex[:63] + "g",
		"sha256:" + helloHex[:62] + " 0",
		" sha256:" + helloHex,
		"sha256:" + helloHex + "\n",
		"SHA256:" + helloHex,
		"sha512:" + helloHex + helloHex,
		"md5:5d41402abc4b2a76b9719d911017c592",
		":" + helloHex,
	}
	for _, s := range invalid {
		if _, err := digest.Parse(s); err == nil {
			t.Errorf("Parse(%q) = nil, want error", s)
		}
	}
}

func TestEqualAfterNormalize(t *testing.T) {
	lower, _ := digest.Parse("sha256:" + helloHex)
	upper, _ := digest.Parse("sha256:" + strings.ToUpper(helloHex))
	if lower != upper {
		t.Errorf("%v != %v, want equal", lower, upper)
	}
}

func TestVerify(t *testing.T) {
	d, err := digest.Parse("sha256:" + strings.ToUpper(helloHex))
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Verify(strings.NewReader("hello")); err != nil {
		t.Errorf("Verify(hello) = %v, want nil", err)
	}

	err = d.Verify(strings.NewReader("hello!"))
	if err == nil {
		t.Fatal("Verify(hello!) = nil, want mismatch")
	}
	for _, want := range []string{"expected sha256:" + helloHex, "got sha256:ce06092fb948d9ffac7d1a376e404b26b7575bcc11ee05a4615fef4fec3a308b"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Verify(hello!) = %q, want it to contain %q", err, want)
		}
	}

	readErr := errors.New("boom")
	if err := d.Verify(iotest.ErrReader(readErr)); !errors.Is(err, readErr) {
		t.Errorf("Verify(failing reader) = %v, want %v", err, readErr)
	}

	if err := (digest.Digest{}).Verify(strings.NewReader("")); err == nil {
		t.Error("zero Digest.Verify = nil, want error")
	}
}
