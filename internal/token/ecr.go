package token

import (
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

const (
	tokenRefreshAfter = 6 * time.Hour // ECR tokens are valid for 12 hours
)

type Token struct {
	Token     string
	ExpiresAt time.Time
	Endpoint  string // ECR endpoint, e.g., "123456789012.dkr.ecr.us-east-1.amazonaws.com"
	Region    string
	Account   string
	Lock      sync.Mutex
}

// NewToken creates a new Token instance with the provided token string and expiry time.
func NewToken(region, account string) *Token {
	t := &Token{
		Region:  region,
		Account: account,
	}
	t.Refresh()
	return t
}

// IsValid checks if the token is still valid based on the current time.
func (t *Token) IsValid() bool {
	return time.Now().Before(t.ExpiresAt) && len(t.Token) > 0
}

// GetToken returns the token string.
func (t *Token) GetToken() string {
	if t.IsExpired() {
		if err := t.Refresh(); err != nil {
			return ""
		}
	}
	return t.Token
}

// GetEndpoint returns the ECR endpoint associated with the token.
func (t *Token) GetEndpoint() string {
	return t.Endpoint
}

// GetExpiresAt returns the expiry time of the token.
func (t *Token) GetExpiresAt() time.Time {
	return t.ExpiresAt
}

// IsExpired checks if the token has expired.
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *Token) Refresh() error {
	t.Lock.Lock()
	defer t.Lock.Unlock()

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(t.Region),
	})
	if err != nil {
		return err
	}

	// Get ECR authorization token
	svc := ecr.New(sess)
	result, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{aws.String(t.Account)},
	})
	if err != nil {
		return err
	}

	if len(result.AuthorizationData) == 0 {
		return nil
	}

	// Update our token and expiry
	t.Token = *result.AuthorizationData[0].AuthorizationToken
	t.ExpiresAt = result.AuthorizationData[0].ExpiresAt.Add(-tokenRefreshAfter)
	t.Endpoint = strings.TrimPrefix(*result.AuthorizationData[0].ProxyEndpoint, "https://")
	return nil
}
