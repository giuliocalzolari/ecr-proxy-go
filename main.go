package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/sethvargo/go-envconfig"
)

const (
	defaultPort       = "5000"
	tokenRefreshAfter = 6 * time.Hour // ECR tokens are valid for 12 hours
)

var (
	awsRegion   string
	ecrEndpoint string
	ecrToken    string
	tokenExpiry time.Time
)

type sysConfig struct {
	Target      string `env:"ECR_TARGET"`
	Region      string `env:"AWS_REGION, default=us-east-1"`
	Account     string `env:"AWS_ACCOUNT_ID"`
	IpWhitelist string `env:"IP_WHITELIST, default="`
	TlsCertFile string `env:"TLS_CERT_FILE, default=/app/tls/tls.crt"`
	TlsKeyFile  string `env:"TLS_KEY_FILE, default=/app/tls/tls.key"`
}

func main() {
	cfg := sysConfig{}
	ctx := context.Background()
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal("%v", err)
	}
	if cfg.Target != "" {
		ecrEndpoint = cfg.Target
		// Try to extract region from ECR_TARGET if possible
		parts := strings.Split(cfg.Target, ".")
		// ECR endpoint format: <account>.dkr.ecr.<region>.amazonaws.com
		if len(parts) >= 6 && parts[2] == "ecr" {
			cfg.Region = parts[3]
			cfg.Account = parts[0]
			log.Printf("Using ECR_TARGET: %s, AWS Region: %s, AWS Account: %s", ecrEndpoint, cfg.Region, cfg.Account)
		} else {
			log.Fatalf("Invalid ECR_TARGET format: %s", cfg.Target)
		}

	} else {

		if cfg.Account == "" {
			// Try to get AWS account ID from STS if not set
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(cfg.Region),
			})
			if err != nil {
				log.Fatalf("Failed to create AWS session: %v", err)
			}
			stsSvc := sts.New(sess)
			idResp, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
			if err != nil || idResp.Account == nil {
				log.Fatal("AWS_ACCOUNT_ID environment variable is required and could not be determined via STS")
			}
			cfg.Account = *idResp.Account
			log.Printf("AWS_ACCOUNT_ID not set, using value from STS: %s", cfg.Account)
		}
		ecrEndpoint = cfg.Account + ".dkr.ecr." + cfg.Region + ".amazonaws.com"
	}

	if _, err := refreshECRToken(cfg); err != nil {
		log.Fatalf("Initial token refresh failed: %v", err)
	}

	// Set up the reverse proxy
	target, _ := url.Parse("https://" + cfg.Target)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = director

	// Get port from environment or use default
	port := defaultPort

	// Set up routes
	http.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		// Check IP whitelist if set
		if cfg.IpWhitelist != "" {
			allowed := isIPAllowed(r.RemoteAddr, cfg.IpWhitelist)
			if !allowed {
				LogEntry(r, "Denied request from IP (not in whitelist)")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		if r.URL.Path != "/v2/" {
			LogEntry(r, "Proxying request to ECR")

		}
		// Refresh token if needed
		if time.Now().After(tokenExpiry) {
			log.Println("ECR token expired or about to expire, refreshing...")
			if _, err := refreshECRToken(cfg); err != nil {
				log.Printf("Failed to refresh ECR token: %v", err)
				http.Error(w, "Failed to refresh ECR token", http.StatusInternalServerError)
				return
			}
			log.Println("ECR token refreshed successfully")
		}
		proxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("Get request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("AWS ECR Proxy is running\n"))
	})

	// TLS configuration
	certFile := cfg.TlsCertFile
	keyFile := cfg.TlsKeyFile
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("TLS cert file not found %s and %s", certFile, keyFile)
	}

	log.Printf("Starting HTTPS ECR proxy on port %s for %s", port, ecrEndpoint)
	log.Fatal(http.ListenAndServeTLS(":"+port, certFile, keyFile, nil))
}

func LogEntry(r *http.Request, msg string) {
	// Log the request details
	logEntry := map[string]interface{}{
		"msg":    msg,
		"method": r.Method,
		"path":   r.URL.Path,
		"remote": r.RemoteAddr,
		"time":   time.Now().Format(time.RFC3339),
	}
	logData, _ := json.Marshal(logEntry)
	log.Println(string(logData))

}

func isIPAllowed(remoteAddr, ipWhitelist string) bool {
	// Split the whitelist into individual CIDRs or IPs
	whitelist := splitAndTrim(ipWhitelist, ",")
	var ipNets []*net.IPNet

	for _, entry := range whitelist {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// If entry is a plain IP, convert to /32 or /128 CIDR
		if !strings.Contains(entry, "/") {
			if strings.Contains(entry, ":") {
				entry += "/128"
			} else {
				entry += "/32"
			}
		}
		_, ipnet, err := net.ParseCIDR(entry)
		if err == nil {
			ipNets = append(ipNets, ipnet)
		}
	}

	// Extract the IP from the remote address, handle [::1]:port and IPv4:port
	ipStr := remoteAddr
	if strings.HasPrefix(ipStr, "[") {
		// IPv6 in [::1]:port format
		if end := strings.LastIndex(ipStr, "]"); end != -1 {
			ipStr = ipStr[1:end]
		}
	} else if colonIndex := strings.LastIndex(ipStr, ":"); colonIndex != -1 {
		ipStr = ipStr[:colonIndex]
	}
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		log.Printf("Failed to parse IP from remoteAddr: %s", remoteAddr)
		return false
	}

	// Check if the IP is in any of the allowed subnets
	for _, ipnet := range ipNets {
		if ipnet.Contains(ip) {
			return true
		}
	}

	log.Printf("IP %s is not in the whitelist", ip)
	return false
}

// splitAndTrim splits a string by the given separator and trims whitespace from each element.
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func director(req *http.Request) {
	// Update request to point to ECR
	req.URL.Scheme = "https"
	req.URL.Host = ecrEndpoint
	req.Host = ecrEndpoint

	// Set the Authorization header with our ECR token
	req.Header.Set("Authorization", "Basic "+ecrToken)
}

func refreshECRToken(cfg sysConfig) (string, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
	})
	if err != nil {
		return "", err
	}

	// Get ECR authorization token
	svc := ecr.New(sess)
	result, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{aws.String(cfg.Account)},
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
