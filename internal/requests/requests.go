package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/merkely-development/reporter/internal/aws"
	"github.com/merkely-development/reporter/internal/kube"
	"github.com/merkely-development/reporter/internal/server"
	"github.com/sirupsen/logrus"
)

// HTTPResponse is a simplified version of http.Response
type HTTPResponse struct {
	Body       string
	StatusCode int
}

// K8sEnvRequest represents the PUT request body to be sent to merkely from k8s
type K8sEnvRequest struct {
	Artifacts []*kube.PodData `json:"artifacts"`
	Type      string          `json:"type"`
	Id        string          `json:"id"`
}

// EcsEnvRequest represents the PUT request body to be sent to merkely from ECS
type EcsEnvRequest struct {
	Artifacts []*aws.EcsTaskData `json:"artifacts"`
	Type      string             `json:"type"`
	Id        string             `json:"id"`
}

// ServerEnvRequest represents the PUT request body to be sent to merkely from a server
type ServerEnvRequest struct {
	Artifacts []*server.ServerData `json:"artifacts"`
	Type      string               `json:"type"`
	Id        string               `json:"id"`
}

func getRetryableHttpClient(maxAPIRetries int, logger *logrus.Logger) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxAPIRetries
	retryClient.Logger = logger
	// get a standard *http.Client from the retryable client
	client := retryClient.StandardClient()
	return client
}

// doRequest sends an HTTP request to a URL and returns the response body and status code
func doRequest(jsonBytes []byte, url, username, password string, maxAPIRetries int, method string, logger *logrus.Logger) (*HTTPResponse, error) {
	client := getRetryableHttpClient(maxAPIRetries, logger)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to create post request to %s : %v", url, err)
	}
	if username == "" {
		username = "unset"
	}
	req.SetBasicAuth(password, username)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := client.Do(req)

	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to send %s request to %s : %v", method, url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to read response from %s request to %s : %v", method, url, err)
	}

	return &HTTPResponse{
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}, nil
}

// SendPayload sends a JSON payload to a URL
func SendPayload(payload interface{}, url, username, token string, maxRetries int, dryRun bool, method string, logger *logrus.Logger) (*HTTPResponse, error) {
	var resp *HTTPResponse
	jsonBytes, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		return resp, err
	}

	if dryRun {
		logger.Info("############### THIS IS A DRY-RUN  ###############")
		logger.Info(string(jsonBytes))
	} else {
		logger.Info("****** Sending the payload to the API ******")
		logger.Info(string(jsonBytes))
		resp, err = doRequest(jsonBytes, url, username, token, maxRetries, method, logger)
		if err != nil {
			return resp, err
		}
		if resp.StatusCode != 201 && resp.StatusCode != 200 {
			return resp, fmt.Errorf("failed to send payload. Got status %d: %v", resp.StatusCode, resp.Body)
		}
	}
	return resp, nil
}
