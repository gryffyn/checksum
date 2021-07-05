package kdf

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"math"
	"reflect"
	"runtime"
	"strings"

	gsk "github.com/gryffyn/go-scrypt-kdf"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

type Params struct {
	Time   uint32
	Memory uint32
	Iter   uint32
	Hmac   func() hash.Hash
	Scrypt gsk.Params
	Cost   int
}

var DefaultParams = Params{
	Time:   3,
	Memory: 32 * 1024,
	Iter:   10000,
	Hmac:   sha256.New,
	Scrypt: gsk.DefaultParams,
	Cost:   10,
}

var SaltLen = 32
var KeyLen = 32

// Argon2i returns Argon2i hash of content in reader
func Argon2i(reader io.Reader, params Params, format bool) (string, error) {
	pw, salt, threads, err := genKDFParams(reader)
	key := argon2.Key(pw, salt, params.Time, params.Memory, threads, uint32(KeyLen))

	if format {
		// Base64 encode the salt and hashed password.
		b64Salt := base64.RawStdEncoding.EncodeToString(salt)
		b64Hash := base64.RawStdEncoding.EncodeToString(key)

		// Return a string using the standard encoded hash representation.
		encodedHash := fmt.Sprintf("$argon2i$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory,
			params.Time, threads, b64Salt, b64Hash)

		return encodedHash, err
	}

	hx := hex.EncodeToString(key)
	return hx, err
}

// Argon2id returns Argon2id hash of content in reader
func Argon2id(reader io.Reader, params Params, format bool) (string, error) {
	pw, salt, threads, err := genKDFParams(reader)
	key := argon2.IDKey(pw, salt, params.Time, params.Memory, threads, uint32(KeyLen))

	if format {
		// Base64 encode the salt and hashed password.
		b64Salt := base64.RawStdEncoding.EncodeToString(salt)
		b64Hash := base64.RawStdEncoding.EncodeToString(key)

		// Return a string using the standard encoded hash representation.
		encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory,
			params.Time, threads, b64Salt, b64Hash)

		return encodedHash, err
	}

	hx := hex.EncodeToString(key)
	return hx, err
}

// PBKDF2 returns PBKDF2 hash of content in reader
func PBKDF2(reader io.Reader, params Params, format bool) (string, error) {
	pw, salt, _, err := genKDFParams(reader)
	key := pbkdf2.Key(pw, salt, int(params.Iter), KeyLen, params.Hmac)

	if format {
		// Base64 encode the salt and hashed password.
		b64Salt := base64.RawStdEncoding.EncodeToString(salt)
		b64Hash := base64.RawStdEncoding.EncodeToString(key)

		// Return a string using the standard encoded hash representation.
		encodedHash := fmt.Sprintf("$pbkdf2-%s$i=%d$%s$%s", getHashFunc(params.Hmac), params.Iter, b64Salt, b64Hash)

		return encodedHash, err
	}

	return string(key), err
}

// ScryptT returns ScryptT hash of content in reader, in tarsnap format
func ScryptT(reader io.Reader, params Params, format bool) (string, error) {
	pw, err := io.ReadAll(reader)
	key, err := gsk.Kdf(pw, params.Scrypt)
	return string(key), err
}

// Scrypt returns Scrypt hash of content in reader
func Scrypt(reader io.Reader, params Params, format bool) (string, error) {
	pw, salt, _, err := genKDFParams(reader)
	s := params.Scrypt
	key, err := scrypt.Key(pw, salt, int(math.Round(math.Pow(2, float64(s.LogN)))), int(s.R), int(s.P), KeyLen)
	return string(key), err
}

// Bcrypt returns Bcrypt hash of content in reader
func Bcrypt(reader io.Reader, params Params, format bool) (string, error) {
	pw, err := io.ReadAll(reader)
	key, err := bcrypt.GenerateFromPassword(pw, params.Cost)
	return string(key), err
}

func genKDFParams(reader io.Reader) ([]byte, []byte, uint8, error) {
	pw, err := io.ReadAll(reader)
	salt := make([]byte, SaltLen)
	_, err = rand.Read(salt)
	if err != nil {
		return []byte{}, []byte{}, 0, err
	}
	threads := runtime.NumCPU()
	if threads > 255 {
		threads = 255
	}
	return pw, salt, uint8(threads), err
}

func getHashFunc(i interface{}) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	fn = strings.TrimPrefix(fn, "crypto/")
	fn = strings.TrimSuffix(fn, ".New")
	return fn
}
