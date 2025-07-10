package main

import (
	"log"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

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
