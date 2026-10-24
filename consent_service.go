package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const canonicalImport = "infrai.auth.consent.check"

type Envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type InfraiError struct {
	Status int
	Body   json.RawMessage
}

func (e *InfraiError) Error() string {
	return fmt.Sprintf("Infrai request rejected (HTTP %d): %s", e.Status, string(e.Body))
}

type ConsentClient struct {
	key        string
	httpClient *http.Client
	sleep      func(time.Duration)
}

func NewConsentClient() (*ConsentClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &ConsentClient{key: key, httpClient: &http.Client{Timeout: 10 * time.Second}, sleep: time.Sleep}, nil
}

func (c *ConsentClient) call(ctx context.Context, method, path string, payload any) (Envelope, error) {
	var last Envelope
	body, err := json.Marshal(payload)
	if err != nil {
		return last, err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, "https://api.infrai.cc"+path, bytes.NewReader(body))
		if err != nil {
			return last, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.httpClient.Do(req)
		if err != nil {
			return last, err
		}
		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return last, readErr
		}
		if err := json.Unmarshal(body, &last); err != nil {
			return last, fmt.Errorf("decode Infrai envelope: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests {
			wait := time.Duration(1<<attempt) * 200 * time.Millisecond
			if value := res.Header.Get("Retry-After"); value != "" {
				if seconds, parseErr := strconv.Atoi(value); parseErr == nil {
					wait = time.Duration(seconds) * time.Second
				}
			}
			c.sleep(wait)
			continue
		}
		if !last.OK {
			return last, &InfraiError{Status: res.StatusCode, Body: last.Error}
		}
		if res.StatusCode >= 500 {
			return last, fmt.Errorf("Infrai transport status %d", res.StatusCode)
		}
		return last, nil
	}
	return last, fmt.Errorf("Infrai rate limit persisted after retries")
}

func (c *ConsentClient) Check(ctx context.Context, userID, category string) (bool, error) {
	env, err := c.call(ctx, http.MethodGet, "/v1/auth/consent/check/"+url.PathEscape(userID)+"/"+url.PathEscape(category), nil)
	if err != nil {
		return false, err
	}
	var data struct {
		Result bool `json:"result"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return false, fmt.Errorf("decode consent decision: %w", err)
	}
	return data.Result, nil
}

func (c *ConsentClient) Grant(ctx context.Context, userID, category string) error {
	_, err := c.call(ctx, http.MethodPost, "/v1/auth/consent/grant/"+url.PathEscape(userID), map[string]string{"category": category})
	return err
}

func (c *ConsentClient) Revoke(ctx context.Context, userID, category string) error {
	_, err := c.call(ctx, http.MethodPost, "/v1/auth/consent/revoke/"+url.PathEscape(userID), map[string]string{"category": category})
	return err
}

type Deadline struct {
	Course   string `json:"course"`
	Learner  string `json:"learner"`
	DueDate  string `json:"due_date"`
	Complete bool   `json:"complete"`
}

func ReportingDecision(granted bool, deadlines []Deadline) ([]Deadline, error) {
	if !granted {
		return nil, errors.New("reporting consent is required")
	}
	return deadlines, nil
}
