package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rk1165/loadsimulator/internal/logger"
)

type OAuthBearer struct {
	ClientId     string
	ClientScope  string
	ClientSecret string
	TokenUrl     string
}

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
}

var client = &http.Client{}

func (o *OAuthBearer) GenerateToken() (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", o.ClientId)
	data.Set("client_secret", o.ClientSecret)
	data.Set("scope", o.ClientScope)

	encodedData := data.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, o.TokenUrl, strings.NewReader(encodedData))

	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.ErrorLog.Printf("OAuthBearer request timed out")
			return "", ctx.Err()
		}
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with statusCode=%d message=%s ", response.StatusCode, response.Status)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var oscarResponse OAuthResponse
	_ = json.Unmarshal(responseBody, &oscarResponse)
	return oscarResponse.AccessToken, nil
}
