// Package util contains utility functions for the edge-tts-go project.
package util

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/difyz9/edge-tts-go/pkg/types"
	"github.com/google/uuid"
)

// ConnectID returns a UUID without dashes.
func ConnectID() string {
	// Generate a random UUID and remove dashes
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// DateToString returns a JavaScript-style date string.
func DateToString() string {
	// Format: "Day Mon DD YYYY HH:MM:SS GMT+0000 (Coordinated Universal Time)"
	return time.Now().UTC().Format("Mon Jan 02 2006 15:04:05 GMT+0000 (Coordinated Universal Time)")
}

// RemoveIncompatibleCharacters removes characters that are not compatible with the TTS service.
func RemoveIncompatibleCharacters(s string) string {
	var result strings.Builder
	for _, r := range s {
		code := int(r)
		if (0 <= code && code <= 8) || (11 <= code && code <= 12) || (14 <= code && code <= 31) {
			result.WriteRune(' ')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GetHeadersAndData returns the headers and data from the given data.
func GetHeadersAndData(data []byte, headerLength int) (map[string]string, []byte) {
	return ProcessWebsocketMessage(data)
}

// ProcessWebsocketMessage parses a websocket message into headers and body.
// It returns a map of headers and the message body as a byte slice.
func ProcessWebsocketMessage(data []byte) (headers map[string]string, body []byte) {
	headers = make(map[string]string)
	
	// Find the end of the headers section
	headerEndIndex := bytes.Index(data, []byte("\r\n\r\n"))
	if headerEndIndex == -1 {
		// If there's no header separator, treat the entire message as body
		return headers, data
	}
	
	// Split headers into individual lines
	headerLines := bytes.Split(data[:headerEndIndex], []byte("\r\n"))
	
	// Parse each header line
	for _, line := range headerLines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			continue
		}
		// Trim any leading/trailing whitespace from the value
		key := string(bytes.TrimSpace(parts[0]))
		value := string(bytes.TrimSpace(parts[1]))
		headers[key] = value
	}
	
	// The body starts after the headers
	body = data[headerEndIndex+4:]
	
	return headers, body
}

// SplitTextByByteLength splits a string into a list of strings of a given byte length.
func SplitTextByByteLength(text string, byteLength int) [][]byte {
	if byteLength <= 0 {
		panic("byteLength must be greater than 0")
	}

	textBytes := []byte(text)
	var result [][]byte

	for len(textBytes) > byteLength {
		// Find the last space in the string
		splitAt := bytes.LastIndex(textBytes[:byteLength], []byte(" "))

		// If no space found, split_at is byteLength
		if splitAt == -1 {
			splitAt = byteLength
		}

		// Verify all & are terminated with a ;
		for bytes.Contains(textBytes[:splitAt], []byte("&")) {
			ampersandIndex := bytes.LastIndex(textBytes[:splitAt], []byte("&"))
			if bytes.Index(textBytes[ampersandIndex:splitAt], []byte(";")) != -1 {
				break
			}

			splitAt = ampersandIndex - 1
			if splitAt < 0 {
				panic("Maximum byte length is too small or invalid text")
			}
			if splitAt == 0 {
				break
			}
		}

		// Append the string to the list
		newText := bytes.TrimSpace(textBytes[:splitAt])
		if len(newText) > 0 {
			result = append(result, newText)
		}
		if splitAt == 0 {
			splitAt = 1
		}
		textBytes = textBytes[splitAt:]
	}

	newText := bytes.TrimSpace(textBytes)
	if len(newText) > 0 {
		result = append(result, newText)
	}

	return result
}

// MkSSML creates a SSML string from the given parameters.
func MkSSML(tc types.TTSConfig, escapedText string) string {
	return fmt.Sprintf(
		"<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'>"+
			"<voice name='%s'>"+
			"<prosody pitch='%s' rate='%s' volume='%s'>"+
			"%s"+
			"</prosody>"+
			"</voice>"+
			"</speak>",
		tc.Voice, tc.Pitch, tc.Rate, tc.Volume, escapedText)
}

// SSMLHeadersPlusData returns the headers and data to be used in the request.
func SSMLHeadersPlusData(requestID, timestamp, ssml string) string {
	return fmt.Sprintf(
		"X-RequestId:%s\r\n"+
			"Content-Type:application/ssml+xml\r\n"+
			"X-Timestamp:%sZ\r\n"+ // This is not a mistake, Microsoft Edge bug.
			"Path:ssml\r\n\r\n"+
			"%s",
		requestID, timestamp, ssml)
}

// CalcMaxMesgSize calculates the maximum message size for the given voice, rate, and volume.
func CalcMaxMesgSize(ttsConfig types.TTSConfig) int {
	websocketMaxSize := 1 << 16
	overheadPerMessage := len(SSMLHeadersPlusData(
		ConnectID(),
		DateToString(),
		MkSSML(ttsConfig, ""),
	)) + 50 // margin of error
	return websocketMaxSize - overheadPerMessage
}

// ValidateStringParam validates the given string parameter based on type and pattern.
func ValidateStringParam(paramName, paramValue, pattern string) (string, error) {
	matched, err := regexp.MatchString(pattern, paramValue)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", fmt.Errorf("invalid %s '%s'", paramName, paramValue)
	}
	return paramValue, nil
}

// ValidateTTSConfig validates the TTSConfig object.
func ValidateTTSConfig(config *types.TTSConfig) error {
	// Possible values for voice are:
	// - Microsoft Server Speech Text to Speech Voice (cy-GB, NiaNeural)
	// - cy-GB-NiaNeural
	// - fil-PH-AngeloNeural
	// Always send the first variant as that is what Microsoft Edge does.
	re := regexp.MustCompile(`^([a-z]{2,})-([A-Z]{2,})-(.+Neural)$`)
	match := re.FindStringSubmatch(config.Voice)
	if match != nil {
		lang := match[1]
		region := match[2]
		name := match[3]
		if strings.Contains(name, "-") {
			region = region + "-" + name[:strings.Index(name, "-")]
			name = name[strings.Index(name, "-")+1:]
		}
		config.Voice = fmt.Sprintf("Microsoft Server Speech Text to Speech Voice (%s-%s, %s)", lang, region, name)
	}

	// Validate the voice, rate, volume, and pitch parameters.
	var err error
	config.Voice, err = ValidateStringParam("voice", config.Voice, `^Microsoft Server Speech Text to Speech Voice \(.+,.+\)$`)
	if err != nil {
		return err
	}
	config.Rate, err = ValidateStringParam("rate", config.Rate, `^[+-]\d+%$`)
	if err != nil {
		return err
	}
	config.Volume, err = ValidateStringParam("volume", config.Volume, `^[+-]\d+%$`)
	if err != nil {
		return err
	}
	config.Pitch, err = ValidateStringParam("pitch", config.Pitch, `^[+-]\d+Hz$`)
	if err != nil {
		return err
	}
	return nil
}

// EscapeXML escapes special characters in XML.
func EscapeXML(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			result.WriteString("&amp;")
		case '<':
			result.WriteString("&lt;")
		case '>':
			result.WriteString("&gt;")
		case '"':
			result.WriteString("&quot;")
		case '\'':
			result.WriteString("&apos;")
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

// IsSpace returns true if the rune is a space.
func IsSpace(r rune) bool {
	return unicode.IsSpace(r)
}
