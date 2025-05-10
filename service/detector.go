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

type Payload struct {
	Image      string   `json:"image"`
	Candidates []string `json:"candidates"`
}

// RequestPayload represents the JSON structure for the API request
type RequestPayload struct {
	Inputs Payload `json:"inputs"`
}

func CreateDetectionPayload(imageFilename string, labelsCSV string) (*RequestPayload, error) {
	// Read the image file
	imageData, err := os.ReadFile(imageFilename)
	if err != nil {
		return nil, fmt.Errorf("error reading image file: %w", err)
	}
	// Convert image data to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)
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
			Image:      base64Data,
			Candidates: labels,
		},
	}
	return payload, nil
}

func RunDetector(payload *RequestPayload, url, apiKey string) error {

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
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		if resp.Body != nil {
			errreason, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Unable to read error response")
				return fmt.Errorf("Request failed with code %d and reason nil", resp.StatusCode)
			}
			return fmt.Errorf("Request failed with code %d and reason %s", resp.StatusCode, string(errreason))
		}
	}
	rawdata, err := io.ReadAll(resp.Body)
	handlers.PanicOnError(err)

	fmt.Println("Response: ", string(rawdata))
	return nil
}
