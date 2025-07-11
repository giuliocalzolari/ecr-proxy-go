package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/giuliocalzolari/ecr-proxy/internal/logx"
	"github.com/giuliocalzolari/ecr-proxy/internal/tls"
	"github.com/giuliocalzolari/ecr-proxy/internal/token"
	"github.com/giuliocalzolari/ecr-proxy/internal/utils"
	"github.com/sethvargo/go-envconfig"
)

const (
	defaultPort       = "5000"
	tokenRefreshAfter = 6 * time.Hour // ECR tokens are valid for 12 hours
)

type sysConfig struct {
	Region      string `env:"AWS_REGION, default=us-east-1"`
	Account     string `env:"AWS_ACCOUNT_ID"`
	IpWhitelist string `env:"IP_WHITELIST, default="`
	TlsCertFile string `env:"TLS_CERT_FILE, default=/tmp/tls.crt"`
	TlsKeyFile  string `env:"TLS_KEY_FILE, default=/tmp/tls.key"`
	Port        string `env:"PORT, default=5000"`
}

func main() {
	cfg := sysConfig{}
	ctx := context.Background()
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatalf("%v", err)
	}

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

	t := token.NewToken(cfg.Region, cfg.Account)

	target, _ := url.Parse("https://" + t.GetEndpoint())
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = t.GetEndpoint()
		req.Host = t.GetEndpoint()
		req.Header.Set("Authorization", "Basic "+t.GetToken())
	}
	// Set up routes
	http.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		// Check IP whitelist if set
		if cfg.IpWhitelist != "" {
			allowed := utils.IsIPAllowed(r.RemoteAddr, cfg.IpWhitelist)
			if !allowed {
				logx.Print(r, "Denied request from IP (not in whitelist)")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		if r.URL.Path != "/v2/" {
			logx.Print(r, "proxy to ECR")
		}
		proxy.ServeHTTP(w, r)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if t.IsValid() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ECR token is not valid or expired"))
		}
	})

	if _, err := os.Stat(cfg.TlsCertFile); os.IsNotExist(err) {
		tls.Generate(cfg.TlsCertFile, cfg.TlsKeyFile)
	}

	log.Printf("Starting HTTPS ECR proxy on port %s for %s", cfg.Port, t.GetEndpoint())
	log.Fatal(http.ListenAndServeTLS(":"+cfg.Port, cfg.TlsCertFile, cfg.TlsKeyFile, nil))
}
