package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

const (
	tfidfThreshold = float64(.08)
	indexThreshold = time.Minute
)

type Entry struct {
	Keywords    []string
	LastServed  time.Time
	LastIndexed time.Time
}

type Store struct {
	Videos   map[string]*Entry // key is VideoTxId
	Channels map[string]*Entry // key is addr
}

type FindResp struct {
	Videos   map[string]int `json:"videos"`
	Channels map[string]int `json:"channels"`
}

var store *Store

func init() {
	store = &Store{
		Videos:   make(map[string]*Entry),
		Channels: make(map[string]*Entry),
	}
}

func find(query string) *FindResp {
	res := &FindResp{
		Videos:   make(map[string]int),
		Channels: make(map[string]int),
	}
	queryKeywords := getKeywords(query)
	for id, video := range store.Videos {
		for _, entryKeyword := range video.Keywords {
			for _, queryKeyword := range queryKeywords {
				if entryKeyword == queryKeyword {
					res.Videos[id] += 1
					video.LastServed = time.Now()
				}
			}
		}
	}
	for id, channel := range store.Channels {
		for _, entryKeyword := range channel.Keywords {
			for _, queryKeyword := range queryKeywords {
				if entryKeyword == queryKeyword {
					res.Channels[id] += 1
					channel.LastServed = time.Now()
				}
			}
		}
	}
	return res
}

func getKeywords(doc string) []string {
	words := make([]string, 0)
	for s, w := range model.Cal(doc) {
		if w > tfidfThreshold {
			words = append(words, strings.ToLower(s))
		}
	}
	return words
}

func calcAddress(owner string) (string, error) {
	n, err := base64.RawURLEncoding.DecodeString(owner)
	if err != nil {
		return "", fmt.Errorf("failed to decode owner: %w", err)
	}
	h := sha256.New()
	h.Write(n)
	sum := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum), nil
}
