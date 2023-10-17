package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/everFinance/goar"
	"github.com/intob/artube-search/cache"
)

var (
	listenAddr string
	gateway    string
	client     *goar.Client
)

func init() {
	listenAddr = os.Getenv("ARTUBE_SEARCH_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":1984"
	}
	fmt.Println("listening on", listenAddr)

	gateway = os.Getenv("ARTUBE_SEARCH_GATEWAY")
	if gateway == "" {
		gateway = "https://arweave.net"
	}
	fmt.Println("using gateway", gateway)

	client = goar.NewClient(gateway)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleFind)
	mux.HandleFunc("/index/video", handleIndexVideo)
	mux.HandleFunc("/index/channel", handleIndexChannel)
	http.ListenAndServe(listenAddr, mux)
}

func handleFind(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	q := r.URL.Query().Get("q")
	if q == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	result, err := cache.Find(q)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to query cache: %s", err)))
		return
	}
	resp, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to marshal response: %s", err)))
		return
	}
	w.Write(resp)
}

func handleIndexVideo(w http.ResponseWriter, r *http.Request) {
	txId := r.URL.Query().Get("txId")
	if txId == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing txId in query"))
		return
	}
	err := indexVideo(txId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("indexing failed: %s", err)))
	}
}

func handleIndexChannel(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("addr")
	if addr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing addr in query"))
		return
	}
	err := indexChannel(addr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("indexing failed: %s", err)))
	}
}

func getTxData(txId string) ([]byte, error) {
	resp, err := http.DefaultClient.Get(fmt.Sprintf("%s/%s", gateway, txId))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
