package s3uploadservice

import (
	"encoding/base64"
	"fmt"
	"go-web-socket/internal/utils/file"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func ReplaceFile(base64Data, fileName string) (string, error) {
	godotenv.Load()

	fileExtension, err := file.GetFileExtensionFromBase64(base64Data)
	if err != nil {
		log.Printf("Error getting file format: %v", err)
		return "", fmt.Errorf("error getting file format")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_DEFAULT_REGION")),
	})
	if err != nil {
		log.Printf("Error creating session: %v", err)
		return "", fmt.Errorf("error creating session: %v", err)
	}

	svc := s3.New(sess)

	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		log.Printf("Error decoding base64: %v", err)
		return "", fmt.Errorf("error decoding base64: %v", err)
	}

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("AWS_BUCKET")),
		Key:         aws.String(fmt.Sprintf("%s.%s", fileName, fileExtension)), // Key is the same as the old file
		Body:        strings.NewReader(string(decodedData)),
		ContentType: aws.String(file.GetContentTypeFromExtension(fileExtension)), // Set MIME type based on file extension
	})

	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return "", fmt.Errorf("error uploading file: %v", err)
	}

	fileURL := fmt.Sprintf("%s.%s", fileName, fileExtension)

	return fileURL, nil
}

func Upload(base64Data string) (string, error) {
	godotenv.Load()

	fileName := uuid.New().String()

	fileExtension, err := file.GetFileExtensionFromBase64(base64Data)
	if err != nil {
		log.Printf("Erro ao obter formato do arquivo: %v", err)
		return "", fmt.Errorf("erro ao obter formato do arquivo")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_DEFAULT_REGION")),
	})
	if err != nil {
		log.Printf("Erro ao criar sessão: %v", err)
		return "", fmt.Errorf("erro ao criar sessão: %v", err)
	}

	svc := s3.New(sess)

	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		log.Printf("Erro ao decodificar base64: %v", err)
		return "", fmt.Errorf("erro ao decodificar base64: %v", err)
	}

	objectKey := fmt.Sprintf("%s.%s", fileName, fileExtension)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("AWS_BUCKET")),
		Key:         aws.String(objectKey),
		Body:        strings.NewReader(string(decodedData)),
		ContentType: aws.String(file.GetContentTypeFromExtension(fileExtension)), // Definir o tipo MIME corretamente
	})

	if err != nil {
		log.Printf("Erro ao fazer upload do arquivo: %v", err)
		return "", fmt.Errorf("erro ao fazer upload do arquivo: %v", err)
	}

	fileURL := fmt.Sprintf("%s.%s", fileName, fileExtension)

	return fileURL, nil
}
