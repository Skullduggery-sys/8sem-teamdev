package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

const appPort = 9000

func Test_PosterScenary(t *testing.T) {
	if err := godotenv.Load(); err != nil {
		t.Errorf("failed to load .env file: %v", err)
	}

	// os.Setenv("APP_HOST", "localhost") // for single testing

	client := &http.Client{}

	signInReq, gotErr := makeSignInRequest()
	assert.NoError(t, gotErr)

	resp, gotErr := client.Do(signInReq)
	assert.NoError(t, gotErr)
	defer resp.Body.Close()

	verify2FAReq, gotErr := makeVerify2FARequest()
	assert.NoError(t, gotErr)

	resp, gotErr = client.Do(verify2FAReq)
	assert.NoError(t, gotErr)
	defer resp.Body.Close()

	token, gotErr := readSignInToken(resp)
	assert.NoError(t, gotErr)
	assert.NotEmpty(t, token)

	createPosterReq, gotErr := makeCreatePosterRequest(token)
	assert.NoError(t, gotErr)

	resp, gotErr = client.Do(createPosterReq)
	assert.NoError(t, gotErr)
	defer resp.Body.Close()

	posterID, gotErr := readPosterID(resp)
	assert.NoError(t, gotErr)

	deleteReq, gotErr := makeDeletePosterRequest(token, posterID)
	assert.NoError(t, gotErr)

	resp, gotErr = client.Do(deleteReq)
	assert.NoError(t, gotErr)
	defer resp.Body.Close()

	signOutReq, gotErr := makeSignOutRequest(token)
	assert.NoError(t, gotErr)

	resp, gotErr = client.Do(signOutReq)
	assert.NoError(t, gotErr)
	defer resp.Body.Close()
}

func makeSignInRequest() (*http.Request, error) {
	host := os.Getenv("APP_HOST")
	if host == "" {
		return nil, fmt.Errorf("APP_HOST is not set")
	}

	url := fmt.Sprintf("http://%s:%d/sign-in", host, appPort)
	rawBody := []byte(`{"name":"alexey vasilyev","login":"alivasilyev","role":"user","email": "convex.hull.trick@mail.ru", "password":"12345"}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(rawBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create sign in request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func makeVerify2FARequest() (*http.Request, error) {
	host := os.Getenv("APP_HOST")
	if host == "" {
		return nil, fmt.Errorf("APP_HOST is not set")
	}

	url := fmt.Sprintf("http://%s:%d/verify-2fa?code=228228&email=convex.hull.trick@mail.ru", host, appPort)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sign in request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func readSignInToken(resp *http.Response) (string, error) {
	type respStruct struct {
		Token string `json:"token"`
	}

	gotResp := &respStruct{}
	err := json.NewDecoder(resp.Body).Decode(gotResp)
	if err != nil {
		return "", err
	}

	return gotResp.Token, nil
}

func makeCreatePosterRequest(token string) (*http.Request, error) {
	host := os.Getenv("APP_HOST")
	if host == "" {
		return nil, fmt.Errorf("APP_HOST is not set")
	}

	url := fmt.Sprintf("http://%s:%d/poster?token=%s", host, appPort, token)
	rawBody := []byte(`{"name":"matrix","year":1999,"userId":1,"genres":["science-fiction","action"],"chrono":136}`)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(rawBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create sign up request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func readPosterID(resp *http.Response) (int, error) {
	type respStruct struct {
		ID int `json:"id"`
	}

	gotResp := &respStruct{}
	err := json.NewDecoder(resp.Body).Decode(gotResp)
	if err != nil {
		return 0, err
	}

	return gotResp.ID, nil
}

func makeDeletePosterRequest(token string, posterID int) (*http.Request, error) {
	host := os.Getenv("APP_HOST")
	if host == "" {
		return nil, fmt.Errorf("APP_HOST is not set")
	}

	url := fmt.Sprintf("http://%s:%d/poster?id=%d&token=%s", host, appPort, posterID, token)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sign up request: %w", err)
	}

	return req, nil
}

func makeSignOutRequest(token string) (*http.Request, error) {
	host := os.Getenv("APP_HOST")
	if host == "" {
		return nil, fmt.Errorf("APP_HOST is not set")
	}

	url := fmt.Sprintf("http://%s:%d/sign-out?token=%s", host, appPort, token)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sign up request: %w", err)
	}

	return req, nil
}
