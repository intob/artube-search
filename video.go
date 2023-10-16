package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func getMetadata(address, videoTxId string) (*Metadata, error) {
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
	`, address, videoTxId))
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

func indexVideo(txId string) error {
	existing := store.Videos[txId]
	if existing != nil && existing.LastIndexed.Add(indexThreshold).After(time.Now()) {
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

	e := &Entry{
		Keywords:    make([]string, 0),
		LastServed:  time.Now(),
		LastIndexed: time.Now(),
	}
	e.Keywords = getKeywords(metadata.Title)
	e.Keywords = append(e.Keywords, getKeywords(metadata.Description)...)

	if len(e.Keywords) == 0 {
		return fmt.Errorf("no keywords above tfidf threshold, will not index")
	}
	store.Videos[txId] = e

	return nil
}
