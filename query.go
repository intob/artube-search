package main

type TxIdQueryResult struct {
	Txs struct {
		Edges []struct {
			Node struct {
				Id string
			} `json:"node"`
		} `json:"edges"`
	} `json:"transactions"`
}
