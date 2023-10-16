package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/everFinance/goar"
	"github.com/wilcosheh/tfidf"
	"github.com/wilcosheh/tfidf/util"
)

var (
	listenAddr string
	gateway    string
	client     *goar.Client
	model      *tfidf.TFIDF
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

	model = tfidf.New()
	lines, err := util.ReadLines("./t8.shakespeare.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, l := range lines {
		model.AddDocs(l)
	}
	fmt.Println("tfidf model ready")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleFind)
	mux.HandleFunc("/index/video", handleIndexVideo)
	mux.HandleFunc("/index/channel", handleIndexChannel)
	http.ListenAndServe(listenAddr, mux)
}

func handleFind(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	results := find(q)
	resp, err := json.Marshal(results)
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
