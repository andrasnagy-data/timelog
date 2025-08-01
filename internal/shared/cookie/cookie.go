package cookie

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const cookieName string = "session"

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

// encrypt creates a tamper-proof session cookie by encrypting the user ID along with
// the cookie name using AES-GCM. Including the cookie name prevents cookie substitution attacks
// where an attacker tries to move cookies between different cookie names.
func encrypt(userId uuid.UUID, secret []byte, cookieName string) (*string, error) {
	// Create a new AES cipher block from the secret key.
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	// Wrap the cipher block in Galois Counter Mode.
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a unique nonce containing 12 random bytes.
	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	// Prepare the plaintext input for encryption. Because we want to
	// authenticate the cookie name as well as the value, we make this plaintext
	// in the format "session:{cookie value}". We use the : character as a
	// separator because it is an invalid character for cookie names and
	// therefore shouldn't appear in them.
	plaintext := fmt.Sprintf("%s:%s", cookieName, userId.String())

	// Encrypt the data using aesGCM.Seal(). By passing the nonce as the first
	// parameter, the encrypted data will be appended to the nonce — meaning
	// that the returned encryptedValue variable will be in the format
	// "{nonce}{encrypted plaintext data}".
	encryptedValue := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64 encode value
	res := base64.URLEncoding.EncodeToString(encryptedValue)
	return &res, nil
}

// decrypt validates and extracts the user ID from a session cookie.
// It verifies both the encrypted content and ensures the cookie name matches expectations,
// preventing cookie substitution attacks and tampering.
func decrypt(encryptedUserId string, secret []byte, expectedCookieName string) (*uuid.UUID, error) {
	// Decode the base64-encoded cookie value. If the cookie didn't contain a
	// valid base64-encoded value, this operation will fail and we return an
	// ErrInvalidValue error.
	value, err := base64.URLEncoding.DecodeString(encryptedUserId)
	if err != nil {
		return nil, ErrInvalidValue
	}

	// Create a new AES cipher block from the secret key.
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	// Wrap the cipher block in Galois Counter Mode.
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size.
	nonceSize := aesGCM.NonceSize()

	// To avoid a potential 'index out of range' panic in the next step, we
	// check that the length of the encrypted value is at least the nonce
	// size.
	if len(value) < nonceSize {
		return nil, ErrInvalidValue
	}

	// Split apart the nonce from the actual encrypted data.
	nonce := value[:nonceSize]
	ciphertext := value[nonceSize:]

	// Use aesGCM.Open() to decrypt and authenticate the data. If this fails,
	// return a ErrInvalidValue error.
	plaintext, err := aesGCM.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return nil, ErrInvalidValue
	}

	// The plaintext value is in the format "{cookie name}:{cookie value}". We
	// use strings.Cut() to split it on the first ":" character.
	actualName, userIDStr, ok := strings.Cut(string(plaintext), ":")
	if !ok {
		return nil, ErrInvalidValue
	}

	// Check that the cookie name is the expected one and hasn't been changed.
	if actualName != expectedCookieName {
		return nil, ErrInvalidValue
	}

	res, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidValue
	}
	return &res, nil
}

func GetCookie(r *http.Request, secret []byte) (*uuid.UUID, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	return decrypt(cookie.Value, secret, cookieName)
}

func SetCookie(w http.ResponseWriter, userID uuid.UUID, secret []byte) error {
	encryptedValue, err := encrypt(userID, secret, cookieName)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    *encryptedValue,
		HttpOnly: true,
		// Send cookie to all routes in the app
		Path:   "/",
		Secure: true,
	})
	return nil
}
