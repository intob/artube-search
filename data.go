package main

import (
	"fmt"
	"io"
	"net/http"
)

func getTxData(txId string) ([]byte, error) {
	resp, err := http.DefaultClient.Get(fmt.Sprintf("%s/%s", gateway, txId))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
