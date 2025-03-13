// Package drm handles DRM operations with clock skew correction.
package drm

import (
	"crypto/sha256"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/difyz9/edge-tts-go/internal/constants"
	"github.com/difyz9/edge-tts-go/pkg/errors"
)

var (
	// clockSkewSeconds is the clock skew in seconds.
	clockSkewSeconds float64

	// mutex is used to protect clockSkewSeconds.
	mutex sync.RWMutex
)

// AdjClockSkewSeconds adjusts the clock skew in seconds in case the system clock is off.
func AdjClockSkewSeconds(skewSeconds float64) {
	mutex.Lock()
	defer mutex.Unlock()
	clockSkewSeconds += skewSeconds
}

// GetUnixTimestamp gets the current timestamp in Unix format with clock skew correction.
func GetUnixTimestamp() float64 {
	mutex.RLock()
	defer mutex.RUnlock()
	return float64(time.Now().UTC().Unix()) + clockSkewSeconds
}

// ParseRFC2616Date parses an RFC 2616 date string into a Unix timestamp.
func ParseRFC2616Date(date string) (float64, error) {
	t, err := time.Parse(time.RFC1123, date)
	if err != nil {
		return 0, err
	}
	return float64(t.UTC().Unix()), nil
}

// HandleClientResponseError handles a client response error.
func HandleClientResponseError(resp *http.Response) error {
	if resp == nil {
		return errors.NewSkewAdjustmentError("no response")
	}

	serverDate := resp.Header.Get("Date")
	if serverDate == "" {
		return errors.NewSkewAdjustmentError("no server date in headers")
	}

	serverDateParsed, err := ParseRFC2616Date(serverDate)
	if err != nil {
		return errors.NewSkewAdjustmentError(fmt.Sprintf("failed to parse server date: %s", serverDate))
	}

	clientDate := GetUnixTimestamp()
	AdjClockSkewSeconds(serverDateParsed - clientDate)
	return nil
}

// GenerateSecMSGEC generates the Sec-MS-GEC token value.
func GenerateSecMSGEC() string {
	// Get the current timestamp in Unix format with clock skew correction
	ticks := GetUnixTimestamp()

	// Switch to Windows file time epoch (1601-01-01 00:00:00 UTC)
	ticks += constants.WinEpoch

	// Round down to the nearest 5 minutes (300 seconds)
	ticks = math.Floor(ticks/300) * 300

	// Convert the ticks to 100-nanosecond intervals (Windows file time format)
	ticks *= constants.SToNS / 100

	// Create the string to hash by concatenating the ticks and the trusted client token
	strToHash := fmt.Sprintf("%.0f%s", ticks, constants.TrustedClientToken)

	// Compute the SHA256 hash and return the uppercased hex digest
	hash := sha256.Sum256([]byte(strToHash))
	return fmt.Sprintf("%X", hash)
}
