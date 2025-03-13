// Package errors contains all the custom error types used in the edge-tts-go project.
package errors

import (
	"errors"
	"fmt"
)

// Base errors
var (
	// ErrEdgeTTS is the base error for the edge-tts-go package.
	ErrEdgeTTS = errors.New("edge-tts error")

	// ErrUnknownResponse is raised when an unknown response is received from the server.
	ErrUnknownResponse = fmt.Errorf("%w: unknown response", ErrEdgeTTS)

	// ErrUnexpectedResponse is raised when an unexpected response is received from the server.
	ErrUnexpectedResponse = fmt.Errorf("%w: unexpected response", ErrEdgeTTS)

	// ErrNoAudioReceived is raised when no audio is received from the server.
	ErrNoAudioReceived = fmt.Errorf("%w: no audio received", ErrEdgeTTS)

	// ErrWebSocketError is raised when a WebSocket error occurs.
	ErrWebSocketError = fmt.Errorf("%w: websocket error", ErrEdgeTTS)

	// ErrSkewAdjustmentError is raised when an error occurs while adjusting the clock skew.
	ErrSkewAdjustmentError = fmt.Errorf("%w: skew adjustment error", ErrEdgeTTS)
)

// NewUnknownResponseError creates a new unknown response error with a custom message.
func NewUnknownResponseError(msg string) error {
	return fmt.Errorf("%w: %s", ErrUnknownResponse, msg)
}

// NewUnexpectedResponseError creates a new unexpected response error with a custom message.
func NewUnexpectedResponseError(msg string) error {
	return fmt.Errorf("%w: %s", ErrUnexpectedResponse, msg)
}

// NewNoAudioReceivedError creates a new no audio received error with a custom message.
func NewNoAudioReceivedError(msg string) error {
	return fmt.Errorf("%w: %s", ErrNoAudioReceived, msg)
}

// NewWebSocketError creates a new WebSocket error with a custom message.
func NewWebSocketError(msg string) error {
	return fmt.Errorf("%w: %s", ErrWebSocketError, msg)
}

// NewSkewAdjustmentError creates a new skew adjustment error with a custom message.
func NewSkewAdjustmentError(msg string) error {
	return fmt.Errorf("%w: %s", ErrSkewAdjustmentError, msg)
}

// IsEdgeTTSError checks if the error is an edge-tts error.
func IsEdgeTTSError(err error) bool {
	return errors.Is(err, ErrEdgeTTS)
}

// IsUnknownResponseError checks if the error is an unknown response error.
func IsUnknownResponseError(err error) bool {
	return errors.Is(err, ErrUnknownResponse)
}

// IsUnexpectedResponseError checks if the error is an unexpected response error.
func IsUnexpectedResponseError(err error) bool {
	return errors.Is(err, ErrUnexpectedResponse)
}

// IsNoAudioReceivedError checks if the error is a no audio received error.
func IsNoAudioReceivedError(err error) bool {
	return errors.Is(err, ErrNoAudioReceived)
}

// IsWebSocketError checks if the error is a WebSocket error.
func IsWebSocketError(err error) bool {
	return errors.Is(err, ErrWebSocketError)
}

// IsSkewAdjustmentError checks if the error is a skew adjustment error.
func IsSkewAdjustmentError(err error) bool {
	return errors.Is(err, ErrSkewAdjustmentError)
}
