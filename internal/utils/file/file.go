package file

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func GetContentTypeFromExtension(extension string) string {
	switch extension {
	case "png":
		return "image/png"
	case "jpeg", "jpg":
		return "image/jpeg"
	case "webp":
		return "image/webp"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func GetFileExtensionFromBase64(base64Str string) (string, error) {
	// Remove the "data:image/*;base64," prefix if present
	if strings.HasPrefix(base64Str, "data:image/") {
		base64Str = strings.Split(base64Str, ",")[1] // Remove the prefix "data:image/*;base64,"
	}

	// Decode the base64 string
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", fmt.Errorf("error decoding base64: %w", err)
	}

	// Detect the MIME type of the file
	mimeType := http.DetectContentType(data)

	// Convert MIME to file extension
	var extension string
	switch mimeType {
	case "image/png":
		extension = "png"
	case "image/jpeg":
		extension = "jpeg"
	case "image/webp":
		extension = "webp"
	case "image/gif":
		extension = "gif"
	default:
		return "", fmt.Errorf("unknown image type: %s", mimeType)
	}

	return extension, nil
}

func GetMimeType(base64Str string) (string, error) {
	// Check if the base64 string is empty
	if strings.TrimSpace(base64Str) == "" {
		log.Println("Error: the base64 string is empty")
		return "", errors.New("the base64 string is empty")
	}

	// Remove the prefix if present
	if strings.HasPrefix(base64Str, "data:image/jpeg;base64,") {
		base64Str = strings.TrimPrefix(base64Str, "data:image/jpeg;base64,")
	} else if strings.HasPrefix(base64Str, "data:image/png;base64,") {
		base64Str = strings.TrimPrefix(base64Str, "data:image/png;base64,")
	}

	// Decode the base64 string
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		log.Printf("Error decoding base64: %v", err)
		return "", fmt.Errorf("error decoding base64: %w", err)
	}

	// Detect the MIME type of the bytes
	mimeType := http.DetectContentType(data)
	return mimeType, nil
}
