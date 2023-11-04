package configfile

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var ErrEmptyHeader = fmt.Errorf("auth header may not be empty")

// CryptPass allows us to validate an input password easily.
type CryptPass string

type AuthType int

const (
	AuthPassword AuthType = iota
	AuthHeader
	AuthNone
)

const (
	authPassword = "!!cryptd!!"
	authHeader   = "webauth"
	authNone     = "noauth"
)

func (c *Config) setupPassword() error {
	pass := c.UIPassword.Val()
	if pass == "" || c.UIPassword.Webauth() {
		return nil
	}

	if !c.UIPassword.IsCrypted() && !strings.Contains(pass, ":") {
		pass = DefaultUsername + ":" + pass
	}

	return c.UIPassword.Set(pass)
}

func (t AuthType) String() string {
	return map[AuthType]string{
		AuthPassword: "Password",
		AuthHeader:   "Header",
		AuthNone:     "No Password",
	}[t]
}

// Set sets an encrypted password.
func (p *CryptPass) Set(pass string) error {
	if strings.HasPrefix(pass, authPassword) ||
		strings.HasPrefix(pass, authHeader+":") ||
		strings.HasPrefix(pass, authNone+":") ||
		pass == "" {
		*p = CryptPass(pass)
		return nil
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("encrypting password: %w", err)
	}

	*p = CryptPass(authPassword + string(bytes))

	return nil
}

func (p *CryptPass) SetNoAuth(header string) error {
	return p.Set(authNone + ":" + header)
}

func (p *CryptPass) SetHeader(header string) error {
	if header == "" {
		return ErrEmptyHeader
	}

	return p.Set(authHeader + ":" + header)
}

// Type returns the authentication type configured.
func (p CryptPass) Type() AuthType {
	switch {
	case p.Noauth():
		return AuthNone
	case p.Webauth():
		return AuthHeader
	default:
		return AuthPassword
	}
}

// Webauth returns true if the password indicates an auth proxy (or no auth) is in use.
func (p CryptPass) Webauth() bool {
	return p == authHeader || strings.HasPrefix(p.Val(), authHeader+":") || p.Noauth()
}

// Header returns the auth proxy header that is configured.
func (p CryptPass) Header() string {
	if split := strings.Split(p.Val(), ":"); len(split) == 2 && split[0] == authHeader {
		return split[1]
	}

	return DefaultHeader
}

// Noauth returns true if the password indicates skipping authentication.
func (p CryptPass) Noauth() bool {
	return p == authNone || strings.HasPrefix(p.Val(), authNone+":")
}

// Val returns the string representation of the current password.
// It may or may not be encrypted.
func (p CryptPass) Val() string {
	return string(p)
}

// Valid checks if a password is valid.
func (p CryptPass) Valid(pass string) bool {
	hash := []byte(strings.TrimPrefix(p.Val(), authPassword))
	return !p.Webauth() && bcrypt.CompareHashAndPassword(hash, []byte(pass)) == nil
}

// IsCrypted checks if a password string is already encrypted.
func (p CryptPass) IsCrypted() bool {
	return strings.HasPrefix(p.Val(), authPassword)
}

// GeneratePassword uses a word list to create a randmo password of two words and a number.
//
//nolint:gosec,gomnd
func GeneratePassword() string {
	title := cases.Title(language.AmericanEnglish)
	pieces := make([]string, 4)

	pieces[0] = words[rand.Intn(len(words))]
	if rand.Intn(10) > 4 {
		pieces[0] = title.String(pieces[0])
	}

	pieces[1] = strconv.Itoa(rand.Intn(89) + 10)
	punctuation := strings.Split(`!@#$%^&*+=/<>|~`, "")
	pieces[2] = punctuation[rand.Intn(len(punctuation))]

	pieces[3] = words[rand.Intn(len(words))]
	if rand.Intn(10) > 4 {
		pieces[3] = title.String(pieces[3])
	}

	rand.Shuffle(len(pieces), func(i, j int) { pieces[i], pieces[j] = pieces[j], pieces[i] })

	return strings.Join(pieces, "")
}
