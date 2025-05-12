package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/paul-at-nangalan/errorhandler/handlers"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type OperationMode string

const (
	OPMODE_TEXT_EMBED  OperationMode = "text-embed"
	OPMODE_IMAGE_EMBED OperationMode = "image-embed"
	OPMODE_MAINOBJECT  OperationMode = "main-object-class"
)

type Payload struct {
	Image      string   `json:"image"`
	Candidates []string `json:"candidates"`
	Type       string   `json:"type"`
	Mode       string   `json:"mode"`
}

// RequestPayload represents the JSON structure for the API request
type RequestPayload struct {
	Inputs Payload `json:"inputs"`
}

func CreateDetectionPayload(imageFilename string, labelsCSV string, mode OperationMode) (*RequestPayload, error) {

	switch mode {
	case OPMODE_IMAGE_EMBED:
		fmt.Println("Creating image request")
		// Read the image file
		imageData, err := os.ReadFile(imageFilename)
		if err != nil {
			return nil, fmt.Errorf("error reading image file: %w", err)
		}
		// Convert image data to base64
		base64Data := base64.StdEncoding.EncodeToString(imageData)
		// Create payload
		payload := &RequestPayload{
			Inputs: Payload{
				Image: base64Data,
				Type:  "get-embeddings",
				Mode:  "image",
			},
		}
		return payload, nil
	case OPMODE_TEXT_EMBED:
		fmt.Println("Creating text request")
		// Split labels string into array
		var labels []string
		if labelsCSV != "" {
			labels = strings.Split(labelsCSV, ",")
			// Trim whitespace from labels
			for i := range labels {
				labels[i] = strings.TrimSpace(labels[i])
			}
		}
		// Create payload
		payload := &RequestPayload{
			Inputs: Payload{
				Candidates: labels,
				Type:       "get-embeddings",
				Mode:       "text",
			},
		}
		return payload, nil
	case OPMODE_MAINOBJECT:
		// Read the image file
		imageData, err := os.ReadFile(imageFilename)
		if err != nil {
			return nil, fmt.Errorf("error reading image file: %w", err)
		}
		// Convert image data to base64
		base64Data := base64.StdEncoding.EncodeToString(imageData)
		// Create payload
		payload := &RequestPayload{
			Inputs: Payload{
				Image: base64Data,
				Type:  "find-main-object",
			},
		}
		return payload, nil
	}
	return nil, fmt.Errorf("Invalid mode %s", mode)
}

func RunDetector(payload *RequestPayload, url, apiKey string) (map[string]interface{}, error) {

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(payload)

	// Create the request
	req, err := http.NewRequest("POST", url, buf)
	handlers.PanicOnError(err)

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		if resp.Body != nil {
			errreason, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Unable to read error response")
				return nil, fmt.Errorf("Request failed with code %d and reason nil", resp.StatusCode)
			}
			return nil, fmt.Errorf("Request failed with code %d and reason %s", resp.StatusCode, string(errreason))
		}
	}
	dec := json.NewDecoder(resp.Body)
	m := make(map[string]interface{})
	err = dec.Decode(&m)

	return m, err
}
