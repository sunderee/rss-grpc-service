package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"rss-grpc/protos"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

const (
	// Using more reliable RSS feeds that are less likely to rate limit
	validRssFeed1  = "https://www.hashicorp.com/blog/feed.xml"
	validRssFeed2  = "https://kubernetes.io/feed.xml"
	invalidRssFeed = "https://invalid.feed.url/rss"
	emptyURL       = ""
)

func setupTestServer() *RssServer {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
	return NewRssServer(parser)
}

func TestGetRssFeed(t *testing.T) {
	server := setupTestServer()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "Valid RSS feed",
			url:         validRssFeed1,
			expectError: false,
		},
		{
			name:        "Empty URL",
			url:         emptyURL,
			expectError: true,
		},
		{
			name:        "Invalid RSS feed",
			url:         invalidRssFeed,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &protos.GetRssFeedRequest{Url: tt.url}
			feed, err := server.GetRssFeed(ctx, req)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			if !assert.NoError(t, err, "Failed to get feed from %s", tt.url) {
				return
			}
			assert.NotNil(t, feed)
			assert.NotEmpty(t, feed.Title)
			assert.NotEmpty(t, feed.Items)
		})
	}
}

func TestGetRssFeeds(t *testing.T) {
	server := setupTestServer()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name        string
		urls        []string
		expectError bool
		minFeeds    int
	}{
		{
			name:        "Multiple valid RSS feeds",
			urls:        []string{validRssFeed1, validRssFeed2},
			expectError: false,
			minFeeds:    1,
		},
		{
			name:        "Empty URL list",
			urls:        []string{},
			expectError: true,
			minFeeds:    0,
		},
		{
			name:        "Mix of valid and invalid feeds",
			urls:        []string{validRssFeed1, invalidRssFeed},
			expectError: false,
			minFeeds:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &protos.GetRssFeedsRequest{Urls: tt.urls}
			feeds, err := server.GetRssFeeds(ctx, req)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			if !assert.NoError(t, err, "Failed to get feeds") {
				t.Logf("Error getting feeds: %v", err)
				return
			}
			assert.NotNil(t, feeds, "Feeds response should not be nil")
			assert.NotEmpty(t, feeds.Feeds, "Feeds list should not be empty")
			assert.GreaterOrEqual(t, len(feeds.Feeds), tt.minFeeds, "Should have at least %d feeds, got %d", tt.minFeeds, len(feeds.Feeds))

			for i, feed := range feeds.Feeds {
				t.Logf("Feed %d: %s", i, feed.Title)
				assert.NotEmpty(t, feed.Title, "Feed %d should have a title", i)
				assert.NotEmpty(t, feed.Items, "Feed %d should have items", i)
			}
		})
	}
}

func TestValidateRssFeed(t *testing.T) {
	server := setupTestServer()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tests := []struct {
		name          string
		url           string
		expectError   bool
		expectedValid bool
	}{
		{
			name:          "Valid RSS feed",
			url:           validRssFeed1,
			expectError:   false,
			expectedValid: true,
		},
		{
			name:          "Empty URL",
			url:           emptyURL,
			expectError:   true,
			expectedValid: false,
		},
		{
			name:          "Invalid RSS feed",
			url:           invalidRssFeed,
			expectError:   false,
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &protos.ValidateRssFeedRequest{Url: tt.url}
			resp, err := server.ValidateRssFeed(ctx, req)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			if !assert.NoError(t, err) {
				return
			}
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectedValid, resp.IsValid)
			assert.Equal(t, tt.url, resp.Url)
		})
	}
}
