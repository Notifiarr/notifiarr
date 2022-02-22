package configfile

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// CryptPass allows us to validate an input password easily.
type CryptPass string

const (
	cryptedPassPfx  = "!!cryptd!!"
	defaultUsername = "admin"
)

func (c *Config) setupPassword() error {
	pass := string(c.UIPassword)
	if pass == "" {
		return nil
	}

	if !c.UIPassword.IsCrypted() && !strings.Contains(pass, ":") {
		pass = defaultUsername + ":" + pass
	}

	if err := c.UIPassword.Set(pass); err != nil {
		return err
	}

	return nil
}

// Set sets a crypted password.
func (p *CryptPass) Set(pass string) error {
	if strings.HasPrefix(pass, cryptedPassPfx) {
		*p = CryptPass(pass)
		return nil
	}

	if pass == "" {
		*p = ""
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("encrypting password: %w", err)
	}

	*p = CryptPass(cryptedPassPfx + string(bytes))

	return nil
}

// Valid checks if a password is valid.
func (p CryptPass) Valid(pass string) bool {
	hash := []byte(strings.TrimPrefix(string(p), cryptedPassPfx))
	return bcrypt.CompareHashAndPassword(hash, []byte(pass)) == nil
}

// IsCrypted checks if a password string is already encrypted.
func (p CryptPass) IsCrypted() bool {
	return strings.HasPrefix(string(p), cryptedPassPfx)
}
