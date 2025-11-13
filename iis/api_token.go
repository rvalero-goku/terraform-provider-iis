package iis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ApiTokenRequest struct {
	ExpiresOn string `json:"expires_on"`
}

type ApiTokenResponse struct {
	AccessToken string `json:"access_token"`
	ID          string `json:"id"`
	ExpiresOn   string `json:"expires_on"`
}

// GenerateApiToken generates an IIS Administration API access token using Windows authentication
// Adapted from Microsoft IIS.Administration utils.ps1 script
func (client Client) GenerateApiToken(ctx context.Context, username, password, domain string) (string, error) {
	// Step 1: GET request to retrieve XSRF token
	getURL := client.Host + "/security/api-keys"
	
	req, err := http.NewRequestWithContext(ctx, "GET", getURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GET request: %w", err)
	}
	
	// Set basic auth with NTLM credentials
	req.SetBasicAuth(formatNTLMUsername(username, domain), password)
	
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get XSRF token: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get XSRF token, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	
	// Extract XSRF token from headers
	xsrfToken := resp.Header.Get("XSRF-TOKEN")
	if xsrfToken == "" {
		return "", fmt.Errorf("XSRF-TOKEN not found in response headers")
	}
	
	// Save cookies from the first request
	cookies := resp.Cookies()
	
	// Step 2: POST request to create API key
	tokenReq := ApiTokenRequest{
		ExpiresOn: "", // Never expires
	}
	
	body, err := json.Marshal(tokenReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}
	
	postURL := client.Host + "/security/api-keys"
	postReq, err := http.NewRequestWithContext(ctx, "POST", postURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create POST request: %w", err)
	}
	
	postReq.SetBasicAuth(formatNTLMUsername(username, domain), password)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("XSRF-TOKEN", xsrfToken)
	
	// Copy cookies from first request
	for _, cookie := range cookies {
		postReq.AddCookie(cookie)
	}
	
	postResp, err := client.HttpClient.Do(postReq)
	if err != nil {
		return "", fmt.Errorf("failed to create API token: %w", err)
	}
	defer postResp.Body.Close()
	
	if postResp.StatusCode != http.StatusCreated && postResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(postResp.Body)
		return "", fmt.Errorf("failed to create API token, status: %d, body: %s", postResp.StatusCode, string(bodyBytes))
	}
	
	var tokenResp ApiTokenResponse
	if err := json.NewDecoder(postResp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}
	
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("access_token not found in response")
	}
	
	// Validate token length (should be 54 characters for IIS Admin API)
	if len(tokenResp.AccessToken) != 54 {
		return "", fmt.Errorf("invalid token length: got %d, expected 54", len(tokenResp.AccessToken))
	}
	
	return tokenResp.AccessToken, nil
}

// formatNTLMUsername formats the username with domain if provided
func formatNTLMUsername(username, domain string) string {
	if domain != "" {
		return domain + "\\" + username
	}
	return username
}
