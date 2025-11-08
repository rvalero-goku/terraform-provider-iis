package iis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func getJson(ctx context.Context, client Client, path string, r interface{}) error {
	data, err := httpGet(ctx, client, path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r)
}

func httpGet(ctx context.Context, client Client, path string) ([]byte, error) {
	response, err := request(ctx, client, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	return fetchBody(response)
}

func httpPost(ctx context.Context, client Client, path string, body interface{}) ([]byte, error) {
	response, err := request(ctx, client, "POST", path, body)
	if err != nil {
		return nil, err
	}
	return fetchBody(response)
}

func httpPatch(ctx context.Context, client Client, path string, body interface{}) ([]byte, error) {
	response, err := request(ctx, client, "PATCH", path, body)
	if err != nil {
		return nil, err
	}
	return fetchBody(response)
}

func httpDelete(ctx context.Context, client Client, path string) error {
	if _, err := request(ctx, client, "DELETE", path, nil); err != nil {
		return err
	}
	return nil
}

func buildRequest(ctx context.Context, client Client, method, path string, body interface{}) (*http.Request, error) {
	b := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(b).Encode(body); err != nil {
			return nil, err
		}
	}

	requestUrl := fmt.Sprintf("%s%s", client.Host, path)
	req, err := http.NewRequestWithContext(ctx, method, requestUrl, b)
	if err != nil {
		return nil, err
	}
	
	// Set authentication and authorization headers
	// Access token is used for API authorization (if available)
	if client.AccessKey != "" {
		req.Header.Set("Access-Token", fmt.Sprintf("Bearer %s", client.AccessKey))
	}
	
	// NTLM authentication: Set basic auth credentials for ntlmssp.Negotiator
	// The ntlmssp.Negotiator transport expects basic auth to be set on requests
	// and will automatically convert them to proper NTLM negotiation
	if client.NTLMUsername != "" && client.NTLMPassword != "" {
		// Format username with domain if provided (domain\username format)
		username := client.NTLMUsername
		if client.NTLMDomain != "" {
			username = fmt.Sprintf("%s\\%s", client.NTLMDomain, client.NTLMUsername)
		}
		req.SetBasicAuth(username, client.NTLMPassword)
	}
	
	// Set required headers for IIS Administration API
	req.Header.Set("Accept", "application/hal+json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func executeRequest(client Client, req *http.Request) (*http.Response, error) {
	// Retry configuration for NTLM authentication issues
	const maxRetries = 3
	const initialBackoff = 500 * time.Millisecond
	
	var response *http.Response
	var err error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 500ms, 1s, 2s
			backoff := initialBackoff * time.Duration(1<<uint(attempt-1))
			time.Sleep(backoff)
		}
		
		response, err = client.HttpClient.Do(req)
		if err != nil {
			// Network errors - retry
			if attempt < maxRetries-1 {
				continue
			}
			return nil, err
		}
		
		// Check if we should retry based on status code
		if shouldRetry(response.StatusCode) && attempt < maxRetries-1 {
			// Close the response body before retrying
			if response.Body != nil {
				response.Body.Close()
			}
			continue
		}
		
		// Success or non-retryable error
		break
	}
	
	if err := guardStatusCode(req.Method, req.URL, response); err != nil {
		return nil, err
	}
	return response, err
}

// shouldRetry determines if a request should be retried based on status code
func shouldRetry(statusCode int) bool {
	// Retry on authentication failures (401), server errors (5xx), and too many requests (429)
	// These are common with NTLM authentication issues and transient server problems
	return statusCode == 401 || statusCode == 429 || (statusCode >= 500 && statusCode < 600)
}

func request(ctx context.Context, client Client, method, path string, body interface{}) (*http.Response, error) {
	req, err := buildRequest(ctx, client, method, path, body)
	if err != nil {
		return nil, err
	}
	return executeRequest(client, req)
}

func fetchBody(res *http.Response) ([]byte, error) {
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err = res.Body.Close(); err != nil {
		return nil, err
	}
	return resBody, nil
}

func guardStatusCode(method string, url *url.URL, response *http.Response) error {
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		var body string
		if buffer, err := fetchBody(response); err == nil {
			body = string(buffer[:])
		}
		return fmt.Errorf("%s %s returned invalid status code: %s\n%s", method, url, response.Status, body)
	}
	return nil
}
