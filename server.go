package main

import (
	"context"
	"log"
	"rss-grpc/protos"
	"time"

	"github.com/mmcdole/gofeed"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RssServer struct {
	protos.UnimplementedRssServiceServer
	parser *gofeed.Parser
}

func NewRssServer(parser *gofeed.Parser) *RssServer {
	log.Println("Initializing the RSS server")
	return &RssServer{parser: parser}
}

func (s *RssServer) GetRssFeed(ctx context.Context, req *protos.GetRssFeedRequest) (*protos.RssFeed, error) {
	if req.Url == "" {
		log.Println("URL cannot be empty")
		return nil, status.Error(codes.InvalidArgument, "URL cannot be empty")
	}

	feed, err := s.parseRSSFeed(req.Url)
	if err != nil {
		log.Printf("Failed to parse RSS feed: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to parse RSS feed: %v", err)
	}

	return feed, nil
}

func (s *RssServer) GetRssFeeds(ctx context.Context, req *protos.GetRssFeedsRequest) (*protos.RssFeeds, error) {
	if len(req.Urls) == 0 {
		log.Println("URL list cannot be empty")
		return nil, status.Error(codes.InvalidArgument, "URL list cannot be empty")
	}

	type result struct {
		index int
		feed  *protos.RssFeed
		err   error
	}

	resultChan := make(chan result, len(req.Urls))
	for i, url := range req.Urls {
		go func(index int, feedUrl string) {
			feed, err := s.parseRSSFeed(feedUrl)
			select {
			case <-ctx.Done():
				return
			case resultChan <- result{index: index, feed: feed, err: err}:
			}
		}(i, url)
	}

	successfulFeeds := make([]*protos.RssFeed, 0, len(req.Urls))
	feedIndexMap := make(map[int]*protos.RssFeed)

	for i := 0; i < len(req.Urls); i++ {
		select {
		case <-ctx.Done():
			return nil, status.Error(codes.DeadlineExceeded, "timeout while fetching feeds")
		case res := <-resultChan:
			if res.err == nil {
				feedIndexMap[res.index] = res.feed
			}
		}
	}

	for i := 0; i < len(req.Urls); i++ {
		if feed, exists := feedIndexMap[i]; exists {
			successfulFeeds = append(successfulFeeds, feed)
		}
	}

	if len(successfulFeeds) == 0 {
		return nil, status.Error(codes.Internal, "failed to parse any of the provided RSS feeds")
	}

	return &protos.RssFeeds{Feeds: successfulFeeds}, nil
}

func (s *RssServer) ValidateRssFeed(ctx context.Context, req *protos.ValidateRssFeedRequest) (*protos.ValidateRssFeedResponse, error) {
	if req.Url == "" {
		log.Println("URL cannot be empty")
		return nil, status.Error(codes.InvalidArgument, "URL cannot be empty")
	}

	feed, err := s.parseRSSFeed(req.Url)
	return &protos.ValidateRssFeedResponse{
		Url:     req.Url,
		IsValid: err == nil && feed != nil,
	}, nil
}

func (s *RssServer) parseRSSFeed(url string) (*protos.RssFeed, error) {
	feed, err := s.parser.ParseURL(url)
	if err != nil {
		log.Printf("Parsing RSS feed failed: %s", err)
		return nil, err
	}

	var description *string
	if feed.Description != "" {
		description = &feed.Description
	}

	var imageURL *string
	if feed.Image != nil {
		imageURL = &feed.Image.URL
	}

	items := make([]*protos.RssFeedItem, len(feed.Items))
	for i, item := range feed.Items {
		var description *string
		if item.Description != "" {
			description = &item.Description
		} else if item.Content != "" {
			description = &item.Content
		}

		var imageURL *string
		if item.Image != nil {
			imageURL = &item.Image.URL
		}

		var date *string
		if item.PublishedParsed != nil {
			parsedDate := item.PublishedParsed.Format(time.RFC3339)
			date = &parsedDate
		}

		items[i] = &protos.RssFeedItem{
			Url:         item.Link,
			Title:       item.Title,
			Description: description,
			ImageUrl:    imageURL,
			Date:        date,
		}
	}

	return &protos.RssFeed{
		Url:         url,
		Title:       feed.Title,
		Description: description,
		ImageUrl:    imageURL,
		Items:       items,
	}, nil
}
