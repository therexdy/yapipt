package internal

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"yapipt/pkg"

	"golang.org/x/crypto/argon2"
)

const (
	memory  = 64 * 1024
	timee    = 3
	threads = 2
	keyLen  = 32
	saltLen = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, timee, memory, threads, keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, timee, threads, b64Salt, b64Hash), nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var m uint32
	var t uint32
	var p uint8

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	hash := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(expected)))

	return subtle.ConstantTimeCompare(hash, expected) == 1, nil
}

const tokenBytes = 32

func (R *Runtime)NewSessionToken(user_name string) (string, error) {
	b := make([]byte, tokenBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	session_token := base64.RawURLEncoding.EncodeToString(b)
	err = R.RedisDB.Set(R.DBContext, user_name, session_token, time.Duration(time.Duration.Hours(24))).Err()
	if err != nil {
		pkg.LogError("Error setting session_token in RedisDB")
	}
	return session_token, nil
}

