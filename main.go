package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type CertUploader struct {
	client *iam.Client
}

func NewCertUploader(cfg aws.Config) *CertUploader {
	return &CertUploader{
		client: iam.NewFromConfig(cfg),
	}
}

func (u *CertUploader) UploadCertificate(ctx context.Context, certName, certPath, keyPath, chainPath string) error {
	// Read certificate file
	certBody, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}

	// Read private key file
	privateKey, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	// Prepare input
	input := &iam.UploadServerCertificateInput{
		ServerCertificateName: aws.String(certName),
		CertificateBody:       aws.String(string(certBody)),
		PrivateKey:            aws.String(string(privateKey)),
	}

	// Add certificate chain if provided
	if chainPath != "" {
		chainBody, err := os.ReadFile(chainPath)
		if err != nil {
			return fmt.Errorf("failed to read certificate chain: %w", err)
		}
		input.CertificateChain = aws.String(string(chainBody))
	}

	// Upload certificate
	output, err := u.client.UploadServerCertificate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload certificate: %w", err)
	}

	fmt.Printf("Certificate uploaded successfully!\n")
	fmt.Printf("Certificate Name: %s\n", *output.ServerCertificateMetadata.ServerCertificateName)
	fmt.Printf("Certificate ID: %s\n", *output.ServerCertificateMetadata.ServerCertificateId)
	fmt.Printf("ARN: %s\n", *output.ServerCertificateMetadata.Arn)

	return nil
}

func (u *CertUploader) ListCertificates(ctx context.Context) error {
	output, err := u.client.ListServerCertificates(ctx, &iam.ListServerCertificatesInput{})
	if err != nil {
		return fmt.Errorf("failed to list certificates: %w", err)
	}

	if len(output.ServerCertificateMetadataList) == 0 {
		fmt.Println("No certificates found")
		return nil
	}

	fmt.Println("Server Certificates:")
	for _, cert := range output.ServerCertificateMetadataList {
		fmt.Printf("  - Name: %s\n", *cert.ServerCertificateName)
		fmt.Printf("    ID: %s\n", *cert.ServerCertificateId)
		fmt.Printf("    ARN: %s\n", *cert.Arn)
		fmt.Printf("    Expiration: %s\n\n", cert.Expiration.String())
	}

	return nil
}

func (u *CertUploader) DeleteCertificate(ctx context.Context, certName string) error {
	input := &iam.DeleteServerCertificateInput{
		ServerCertificateName: aws.String(certName),
	}

	_, err := u.client.DeleteServerCertificate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	fmt.Printf("Certificate '%s' deleted successfully!\n", certName)
	return nil
}

func main() {
	// Define flags
	certName := flag.String("name", "", "Certificate name (required for upload/delete)")
	certPath := flag.String("cert", "", "Path to certificate file (required for upload)")
	keyPath := flag.String("key", "", "Path to private key file (required for upload)")
	chainPath := flag.String("chain", "", "Path to certificate chain file (optional)")
	listCerts := flag.Bool("list", false, "List existing certificates")
	deleteCert := flag.Bool("delete", false, "Delete a certificate")
	region := flag.String("region", "us-east-1", "AWS region")

	flag.Parse()

	// Load AWS configuration
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(*region))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load AWS config: %v\n", err)
		os.Exit(1)
	}

	uploader := NewCertUploader(cfg)

	// Handle list command
	if *listCerts {
		if err := uploader.ListCertificates(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle delete command
	if *deleteCert {
		if *certName == "" {
			fmt.Fprintf(os.Stderr, "Error: -name is required for delete\n\n")
			flag.Usage()
			os.Exit(1)
		}
		if err := uploader.DeleteCertificate(ctx, *certName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Validate required flags for upload
	if *certName == "" || *certPath == "" || *keyPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -name, -cert, and -key are required for upload\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Upload certificate
	if err := uploader.UploadCertificate(ctx, *certName, *certPath, *keyPath, *chainPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
