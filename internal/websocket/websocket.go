// Package websocket handles the WebSocket communication with the TTS service.
package websocket

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/difyz9/edge-tts-go/internal/constants"
	"github.com/difyz9/edge-tts-go/internal/drm"
	"github.com/difyz9/edge-tts-go/pkg/errors"
	"github.com/difyz9/edge-tts-go/pkg/types"
	"github.com/difyz9/edge-tts-go/pkg/util"
	"github.com/gorilla/websocket"
)

// Client is a WebSocket client for the TTS service.
type Client struct {
	conn           *websocket.Conn
	proxy          string
	connectTimeout int
	receiveTimeout int
}

// NewClient creates a new WebSocket client.
func NewClient(proxy string, connectTimeout, receiveTimeout int) *Client {
	return &Client{
		proxy:          proxy,
		connectTimeout: connectTimeout,
		receiveTimeout: receiveTimeout,
	}
}

// Connect connects to the TTS service.
func (c *Client) Connect(ctx context.Context) error {
	// Parse the WebSocket URL
	u, err := url.Parse(constants.WSSURL + "&Sec-MS-GEC=" + drm.GenerateSecMSGEC() +
		"&Sec-MS-GEC-Version=" + constants.SecMSGECVersion +
		"&ConnectionId=" + util.ConnectID())
	if err != nil {
		return err
	}

	// Create dialer with compression enabled (equivalent to compress=15 in Python version)
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 0, // No timeout
		EnableCompression: true,
	}

	// Set proxy if provided
	if c.proxy != "" {
		proxyURL, err := url.Parse(c.proxy)
		if err != nil {
			return err
		}
		dialer.Proxy = http.ProxyURL(proxyURL)
	}

	// Set headers
	header := http.Header{}
	for k, v := range constants.WSSHeaders {
		header.Set(k, v)
	}

	// Connect to the WebSocket server
	conn, resp, err := dialer.DialContext(ctx, u.String(), header)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusForbidden {
			err = drm.HandleClientResponseError(resp)
			if err != nil {
				return err
			}

			// Retry the connection
			u, err = url.Parse(constants.WSSURL + "&Sec-MS-GEC=" + drm.GenerateSecMSGEC() +
				"&Sec-MS-GEC-Version=" + constants.SecMSGECVersion +
				"&ConnectionId=" + util.ConnectID())
			if err != nil {
				return err
			}

			conn, _, err = dialer.DialContext(ctx, u.String(), header)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Enable compression (equivalent to compress=15 in Python version)
	conn.EnableWriteCompression(true)

	c.conn = conn
	return nil
}

// Close closes the WebSocket connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendCommandRequest sends the command request to the service.
func (c *Client) SendCommandRequest() error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	message := fmt.Sprintf(
		"X-Timestamp:%s\r\n"+
			"Content-Type:application/json; charset=utf-8\r\n"+
			"Path:speech.config\r\n\r\n"+
			`{"context":{"synthesis":{"audio":{"metadataoptions":{`+
			`"sentenceBoundaryEnabled":"false","wordBoundaryEnabled":"true"},`+
			`"outputFormat":"audio-24khz-48kbitrate-mono-mp3"`+
			`}}}}`,
		util.DateToString())

	return c.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// SendSSMLRequest sends the SSML request to the service.
func (c *Client) SendSSMLRequest(partialText []byte, ttsConfig types.TTSConfig) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	message := util.SSMLHeadersPlusData(
		util.ConnectID(),
		util.DateToString(),
		util.MkSSML(ttsConfig, string(partialText)),
	)

	return c.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// ReceiveMessage receives a message from the service.
func (c *Client) ReceiveMessage() (types.TTSChunk, error) {
	if c.conn == nil {
		return types.TTSChunk{}, fmt.Errorf("not connected")
	}

	messageType, data, err := c.conn.ReadMessage()
	if err != nil {
		return types.TTSChunk{}, errors.NewWebSocketError(err.Error())
	}

	switch messageType {
	case websocket.TextMessage:
		headers, messageData := util.ProcessWebsocketMessage(data)

		path := headers["Path"]
		if path == "audio.metadata" {
			// Parse the metadata and return it
			return c.parseMetadata(messageData)
		} else if path == "turn.end" {
			// Return a special chunk to indicate the end of the turn
			return types.TTSChunk{Type: "turn.end"}, nil
		} else if path != "response" && path != "turn.start" {
			return types.TTSChunk{}, errors.NewUnknownResponseError("unknown path received: " + path)
		}

		// For response and turn.start, just return an empty chunk
		return types.TTSChunk{Type: path}, nil

	case websocket.BinaryMessage:
		// Message is too short to contain header length
		if len(data) < 2 {
			return types.TTSChunk{}, errors.NewUnexpectedResponseError("binary message is too short")
		}

		// The first two bytes of the binary message contain the header length
		headerLength := int(binary.BigEndian.Uint16(data[:2]))
		if headerLength+2 > len(data) {
			return types.TTSChunk{}, errors.NewUnexpectedResponseError("header length is greater than the length of the data")
		}

		// Extract the audio data directly
		// The audio data starts after the header (2 bytes for length + headerLength)
		audioBinaryData := data[headerLength+2:]
		
		// Process the headers to check if this is audio data
		headerData := data[2:2+headerLength]
		headers := make(map[string]string)
		for _, line := range bytes.Split(headerData, []byte("\r\n")) {
			if len(line) == 0 {
				continue
			}
			parts := bytes.SplitN(line, []byte(":"), 2)
			if len(parts) != 2 {
				continue
			}
			headers[string(parts[0])] = string(bytes.TrimSpace(parts[1]))
		}
		
		// Check if the path is audio
		pathHeader, exists := headers["Path"]
		if !exists || pathHeader != "audio" {
			return types.TTSChunk{}, errors.NewUnexpectedResponseError("received binary message, but the path is not audio")
		}
		
		// Check content type
		contentType, hasContentType := headers["Content-Type"]
		if !hasContentType {
			// No Content-Type header
			if len(audioBinaryData) == 0 {
				return types.TTSChunk{Type: "audio", Data: []byte{}}, nil
			}
			// If the data is not empty, then we need to raise an exception
			return types.TTSChunk{}, errors.NewUnexpectedResponseError("received binary message with no Content-Type, but with data")
		}
		
		// Has Content-Type header
		if contentType != "audio/mpeg" {
			return types.TTSChunk{}, errors.NewUnexpectedResponseError("received binary message, but with an unexpected Content-Type: " + contentType)
		}
		
		// If the data is empty now, then we need to raise an exception
		if len(audioBinaryData) == 0 {
			return types.TTSChunk{}, errors.NewUnexpectedResponseError("received binary message, but it is missing the audio data")
		}
		
		// Return the audio data
		return types.TTSChunk{Type: "audio", Data: audioBinaryData}, nil

	default:
		return types.TTSChunk{}, errors.NewUnexpectedResponseError("unexpected message type")
	}
}

// parseMetadata parses the metadata from the message data.
func (c *Client) parseMetadata(data []byte) (types.TTSChunk, error) {
	var jsonData map[string]interface{}
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return types.TTSChunk{}, err
	}

	metadata, ok := jsonData["Metadata"].([]interface{})
	if !ok {
		return types.TTSChunk{}, errors.NewUnexpectedResponseError("invalid metadata format")
	}

	for _, metaObj := range metadata {
		metaMap, ok := metaObj.(map[string]interface{})
		if !ok {
			continue
		}

		metaType, ok := metaMap["Type"].(string)
		if !ok {
			continue
		}

		if metaType == "WordBoundary" {
			metaData, ok := metaMap["Data"].(map[string]interface{})
			if !ok {
				continue
			}

			offset, ok := metaData["Offset"].(float64)
			if !ok {
				continue
			}

			duration, ok := metaData["Duration"].(float64)
			if !ok {
				continue
			}

			textData, ok := metaData["text"].(map[string]interface{})
			if !ok {
				continue
			}

			text, ok := textData["Text"].(string)
			if !ok {
				continue
			}

			return types.TTSChunk{
				Type:     metaType,
				Offset:   offset,
				Duration: duration,
				Text:     text,
			}, nil
		}

		if metaType == "SessionEnd" {
			continue
		}

		return types.TTSChunk{}, errors.NewUnknownResponseError("unknown metadata type: " + metaType)
	}

	return types.TTSChunk{}, errors.NewUnexpectedResponseError("no WordBoundary metadata found")
}
