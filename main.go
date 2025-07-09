package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

const (
	defaultPort       = "5000"
	tokenRefreshAfter = 6 * time.Hour // ECR tokens are valid for 12 hours
)

var (
	awsRegion   string
	awsAccount  string
	ecrEndpoint string
	ecrToken    string
	tokenExpiry time.Time
)

func main() {
	// Load configuration
	awsRegion = os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1" // Default region if not set
		log.Printf("AWS_REGION not set, using default: %s", awsRegion)
	}

	awsAccount = os.Getenv("AWS_ACCOUNT_ID")
	if awsAccount == "" {
		log.Fatal("AWS_ACCOUNT_ID environment variable is required")
	}

	ecrEndpoint = awsAccount + ".dkr.ecr." + awsRegion + ".amazonaws.com"

	// Initialize the first token
	if _, err := refreshECRToken(); err != nil {
		log.Fatalf("Initial token refresh failed: %v", err)
	}

	// Set up the reverse proxy
	target, _ := url.Parse("https://" + ecrEndpoint)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = director

	// Get port from environment or use default
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = defaultPort
	}

	// Set up routes
	http.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxying request: %s %s", r.Method, r.URL.Path)
		// Refresh token if needed
		if time.Now().After(tokenExpiry) {
			log.Println("ECR token expired or about to expire, refreshing...")
			if _, err := refreshECRToken(); err != nil {
				log.Printf("Failed to refresh ECR token: %v", err)
				http.Error(w, "Failed to refresh ECR token", http.StatusInternalServerError)
				return
			}
			log.Println("ECR token refreshed successfully")
		}
		proxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Get request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("AWS ECR Proxy is running\n"))
	})

	// TLS configuration
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile == "" {
		certFile = "server.crt" // Default certificate file
	}
	if keyFile == "" {
		keyFile = "server.key" // Default key file
	}
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Printf("TLS cert file not found, generating self-signed certificate at %s and %s", certFile, keyFile)
		err := generateSelfSignedCert(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed to generate self-signed certificate: %v", err)
		}
	}

	log.Printf("Starting HTTPS ECR proxy on port %s for %s", port, ecrEndpoint)
	log.Fatal(http.ListenAndServeTLS(":"+port, certFile, keyFile, nil))
}

func generateSelfSignedCert(certFile, keyFile string) error {
	log.Printf("Generating self-signed certificate at %s and %s", certFile, keyFile)

	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Self-Signed Certificate"},
			CommonName:   "localhost",
		},
		// DNSNames:              []string{"localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// Create certificate file
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}); err != nil {
		return err
	}

	// Create key file
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}); err != nil {
		return err
	}

	return nil
}

func director(req *http.Request) {
	// Update request to point to ECR
	req.URL.Scheme = "https"
	req.URL.Host = ecrEndpoint
	req.Host = ecrEndpoint

	// Set the Authorization header with our ECR token
	req.Header.Set("Authorization", "Basic "+ecrToken)
}

func refreshECRToken() (string, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	if err != nil {
		return "", err
	}

	// Get ECR authorization token
	svc := ecr.New(sess)
	result, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{aws.String(awsAccount)},
	})
	if err != nil {
		return "", err
	}

	if len(result.AuthorizationData) == 0 {
		return "", nil
	}

	// Update our token and expiry
	ecrToken = *result.AuthorizationData[0].AuthorizationToken
	tokenExpiry = result.AuthorizationData[0].ExpiresAt.Add(-tokenRefreshAfter)

	log.Println("Successfully refreshed ECR authorization token")
	return ecrToken, nil
}
