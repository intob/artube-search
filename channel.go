package main

import (
	"encoding/json"
	"fmt"

	"github.com/intob/artube-search/cache"
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

func getAvatarTxId(addr string) (string, error) {
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
					values: ["avatar"]
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
	`, addr))
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}
	result := &TxIdQueryResult{}
	err = json.Unmarshal(resp, result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal query result: %w", err)
	}
	if len(result.Txs.Edges) == 0 {
		return "", fmt.Errorf("no channel tx found")
	}
	return result.Txs.Edges[0].Node.Id, nil
}

func indexChannel(addr string) error {
	if !cache.ShouldIndex(addr) {
		return fmt.Errorf("already indexed within threshold")
	}

	channel, err := getChannel(addr)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	avatarTxId, err := getAvatarTxId(addr)
	if err != nil {
		return fmt.Errorf("failed to get avatarTxId: %w", err)
	}

	return cache.IndexChannel(addr, &cache.ChannelContent{
		Name:        channel.Name,
		Description: channel.Description,
		AvatarTxId:  avatarTxId,
	})
}
