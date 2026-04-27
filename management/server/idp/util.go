package idp

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

type JsonParser struct{}

func (JsonParser) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JsonParser) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// GeneratePassword generates user password
func GeneratePassword(passwordLength, minSpecialChar, minNum, minUpperCase int) (string, error) {
	var password strings.Builder

	//Set special character
	for i := 0; i < minSpecialChar; i++ {
		random, err := secureRandomIndex(len(specialCharSet))
		if err != nil {
			return "", err
		}
		password.WriteString(string(specialCharSet[random]))
	}

	//Set numeric
	for i := 0; i < minNum; i++ {
		random, err := secureRandomIndex(len(numberSet))
		if err != nil {
			return "", err
		}
		password.WriteString(string(numberSet[random]))
	}

	//Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random, err := secureRandomIndex(len(upperCharSet))
		if err != nil {
			return "", err
		}
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random, err := secureRandomIndex(len(allCharSet))
		if err != nil {
			return "", err
		}
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	for i := len(inRune) - 1; i > 0; i-- {
		j, err := secureRandomIndex(i + 1)
		if err != nil {
			return "", err
		}
		inRune[i], inRune[j] = inRune[j], inRune[i]
	}
	return string(inRune), nil
}

func secureRandomIndex(bound int) (int, error) {
	if bound <= 0 {
		return 0, fmt.Errorf("invalid random bound %d", bound)
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(bound)))
	if err != nil {
		return 0, fmt.Errorf("secure random index: %w", err)
	}
	return int(n.Int64()), nil
}

// baseURL returns the base url  by concatenating
// the scheme and host components of the parsed URL.
func baseURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsedURL.Scheme + "://" + parsedURL.Host
}

const (
	// Provides the env variable name for use with idpTimeout function
	idpTimeoutEnv = "NB_IDP_TIMEOUT"
	// Sets the defaultTimeout to 10s.
	defaultTimeout = 10 * time.Second
)

// idpTimeout returns a timeout value for the IDP
func idpTimeout() time.Duration {
	timeoutStr, ok := os.LookupEnv(idpTimeoutEnv)
	if !ok || timeoutStr == "" {
		return defaultTimeout
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return defaultTimeout
	}
	return timeout
}
