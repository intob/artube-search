package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/intob/artube-search/cache"
)

type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func getMetadata(addr, videoTxId string) (*Metadata, error) {
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
					values: ["metadata"]
				},
				{
					name: "Artube-Video",
					values: ["%s"]
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
	`, addr, videoTxId))
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	result := &TxIdQueryResult{}
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal query result: %w", err)
	}
	if len(result.Txs.Edges) == 0 {
		return nil, fmt.Errorf("no metadata tx found")
	}
	metadataResp, err := getTxData(result.Txs.Edges[0].Node.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}
	metadata := &Metadata{}
	return metadata, json.Unmarshal(metadataResp, metadata)
}

func getPosterTxId(addr, videoTxId string) (string, error) {
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
					values: ["poster"]
				},
				{
					name: "Artube-Video",
					values: ["%s"]
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
	`, addr, videoTxId))
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}
	result := &TxIdQueryResult{}
	err = json.Unmarshal(resp, result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal query result: %w", err)
	}
	if len(result.Txs.Edges) == 0 {
		return "", fmt.Errorf("no poster tx found")
	}
	return result.Txs.Edges[0].Node.Id, nil
}

func indexVideo(txId string) error {
	if !cache.ShouldIndex(txId) {
		return fmt.Errorf("already indexed within threshold")
	}

	tx, err := client.GetTransactionByID(txId)
	if err != nil {
		return err
	}
	addr, err := calcAddress(tx.Owner)
	if err != nil {
		return fmt.Errorf("failed to calculate address: %w", err)
	}

	metadata, err := getMetadata(addr, txId)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	posterTxId, err := getPosterTxId(addr, txId)
	if err != nil {
		return fmt.Errorf("failed to get posterTxId: %w", err)
	}

	return cache.IndexVideo(txId, &cache.VideoContent{
		Title:       metadata.Title,
		Description: metadata.Description,
		PosterTxId:  posterTxId,
	})
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
