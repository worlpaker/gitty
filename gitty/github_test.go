package gitty

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"

	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	// Must be same as token const key.
	tokenKey := "GH_TOKEN"
	fakeValue := gofakeit.LoremIpsumWord()

	tests := []struct {
		name     string
		auth     bool
		expected *github.Client
	}{
		{
			name:     "not authorized client",
			expected: github.NewClient(nil),
		},
		{
			name:     "authorized client",
			auth:     true,
			expected: github.NewClient(nil).WithAuthToken(fakeValue),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.auth {
				err := os.Unsetenv(tokenKey)
				assert.Nil(t, err)
				client := newClient()
				assert.Equal(t, test.expected, client)
			} else {
				err := os.Setenv(tokenKey, fakeValue)
				assert.Nil(t, err)
				t.Cleanup(func() {
					err = os.Unsetenv(tokenKey)
					assert.Nil(t, err)
				})
				client := newClient()
				assert.Equal(t, test.expected.UserAgent, client.UserAgent)
			}
		})
	}
}

var mockGetBody = []byte(`{"data":"test"}`)

type mock struct{}

func (m mock) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(mockGetBody)),
	}
	return resp, nil
}

func setup() *service {
	mockClient := &http.Client{
		Transport: mock{},
	}
	s := &service{
		client: github.NewClient(mockClient),
	}
	return s
}

func TestGet(t *testing.T) {
	s := setup()
	resp, err := s.Get("https://test.com")
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	expectedBody, _ := io.ReadAll(resp.Body)
	assert.Equal(t, mockGetBody, expectedBody)
}

func TestGetContents(t *testing.T) {
	s := setup()
	opts := &github.RepositoryContentGetOptions{
		Ref: "main",
	}
	_, _, resp, err := s.GetContents(context.Background(), "owner", "repo", "path", opts)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRateLimit(t *testing.T) {
	s := setup()
	_, resp, err := s.RateLimit(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUser(t *testing.T) {
	s := setup()
	_, resp, err := s.GetUser(context.Background(), "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
