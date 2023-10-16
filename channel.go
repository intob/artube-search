package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Channel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func getChannel(address string) (*Channel, error) {
	resp, err := client.GraphQL(fmt.Sprintf(`
	{
		transactions(
			owners: ["%s"]
			tags: [
				{
					name: "App-Name",
					values: ["artube"]
				},
				{
					name: "Artube-Type",
					values: ["channel"]
				}
			]
		) {
			edges {
				node {
					id
				}
			}
		}
	}
	`, address))
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	result := &TxIdQueryResult{}
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal query result: %w", err)
	}
	if len(result.Txs.Edges) == 0 {
		return nil, fmt.Errorf("no channel tx found")
	}
	channelResp, err := getTxData(result.Txs.Edges[0].Node.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel: %w", err)
	}
	channel := &Channel{}
	return channel, json.Unmarshal(channelResp, channel)
}

func indexChannel(addr string) error {
	existing := store.Channels[addr]
	if existing != nil && existing.LastIndexed.Add(indexThreshold).After(time.Now()) {
		return fmt.Errorf("already indexed within threshold")
	}
	channel, err := getChannel(addr)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	e := &Entry{
		Keywords:    make([]string, 0),
		LastServed:  time.Now(),
		LastIndexed: time.Now(),
	}
	e.Keywords = getKeywords(channel.Name)
	e.Keywords = append(e.Keywords, getKeywords(channel.Description)...)

	if len(e.Keywords) == 0 {
		return fmt.Errorf("no keywords above tfidf threshold, will not index")
	}
	store.Channels[addr] = e

	return nil
}
