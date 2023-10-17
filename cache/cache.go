package cache

import (
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/wilcosheh/tfidf"
	"github.com/wilcosheh/tfidf/util"
)

const (
	tfidfThreshold = float64(.08)
	indexThreshold = time.Minute
)

type Entry struct {
	Keywords    []string    `json:"-"`
	LastIndexed time.Time   `json:"-"`
	Content     interface{} `json:"content"`
	Type        string      `json:"type"`
}

type VideoContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Likes       int    `json:"likes"`
	PosterTxId  string `json:"posterTxId"`
}

type ChannelContent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarTxId  string `json:"avatarTxId"`
}

type FindResult struct {
	Items map[string]*FoundItem `json:"items"`
}

type FoundItem struct {
	Score int    `json:"score"`
	Entry *Entry `json:"entry"`
}

var (
	model *tfidf.TFIDF
	dir   string
)

func init() {
	dir = os.Getenv("ARTUBE_SEARCH_CACHE_DIR")
	if dir == "" {
		dir = "/tmp/artube-search"
	}
	fmt.Println("cache dir", dir)

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}

	model = tfidf.New()
	lines, err := util.ReadLines("assets/t8.shakespeare.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, l := range lines {
		model.AddDocs(l)
	}
	fmt.Println("tfidf model ready")

	gob.Register(&VideoContent{})
	gob.Register(&ChannelContent{})
}

func Find(query string) (*FindResult, error) {
	res := &FindResult{
		Items: make(map[string]*FoundItem),
	}
	queryKeywords := getKeywords(query)
	keys, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	for _, entry := range keys {
		key := entry.Name()
		f, err := os.Open(path.Join(dir, key))
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		entry := &Entry{}
		dec := gob.NewDecoder(f)
		err = dec.Decode(entry)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to decode entry: %w", err)
		}
		for _, entryKeyword := range entry.Keywords {
			for _, queryKeyword := range queryKeywords {
				if entryKeyword == queryKeyword {
					if res.Items[key] == nil {
						res.Items[key] = &FoundItem{
							Entry: entry,
						}
					}
					res.Items[key].Score += 1
				}
			}
		}
	}
	return res, nil
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

func IndexVideo(txId string, content *VideoContent) error {
	e := &Entry{
		Keywords:    make([]string, 0),
		LastIndexed: time.Now(),
		Content:     content,
		Type:        "video",
	}
	e.Keywords = getKeywords(content.Title)
	e.Keywords = append(e.Keywords, getKeywords(content.Description)...)

	if len(e.Keywords) == 0 {
		return fmt.Errorf("no keywords above tfidf threshold, will not index")
	}

	return writeFile(txId, e)
}

func IndexChannel(addr string, content *ChannelContent) error {
	e := &Entry{
		Keywords:    make([]string, 0),
		LastIndexed: time.Now(),
		Content:     content,
		Type:        "channel",
	}
	e.Keywords = getKeywords(content.Name)
	e.Keywords = append(e.Keywords, getKeywords(content.Description)...)

	if len(e.Keywords) == 0 {
		return fmt.Errorf("no keywords above tfidf threshold, will not index")
	}

	return writeFile(addr, e)
}

func ShouldIndex(key string) bool {
	entry, err := readFile(key)
	if err != nil {
		return true
	}
	if entry.LastIndexed.Add(indexThreshold).Before(time.Now()) {
		return true
	}
	return false
}

func readFile(key string) (*Entry, error) {
	f, err := os.Open(path.Join(dir, key))
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	entry := &Entry{}
	err = dec.Decode(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to decode entry gob: %w", err)
	}
	return entry, nil
}

func writeFile(key string, entry *Entry) error {
	f, err := os.Create(path.Join(dir, key))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(entry)
	if err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}
	return nil
}
