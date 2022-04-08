package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	elastic "github.com/elastic/go-elasticsearch/v8"
)

// Client wraps elasticsearch client.
type Client struct {
	es *elastic.Client
}

func newClient() (Client, error) {
	// Initialize a client with the default settings.
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
		return Client{}, err
	}
	c := Client{es: es}

	return c, nil
}

func (c *Client) getInfo() error {
	var r map[string]interface{}
	// Get the information of elasticsearch cluster.
	res, err := c.es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
		return err
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
		return err
	}
	// Print client and server version numbers.
	log.Printf("Client: %s", elasticsearch.Version)
	log.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
	log.Println(strings.Repeat("~", 37))
	return nil
}

func (c *Client) indexDocument(index string, document string, id int) error {
	// Build the request body.
	var b strings.Builder
	b.WriteString(`{"title" : "`)
	b.WriteString(document)
	b.WriteString(`"}`)

	// Set up the request object.
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: strconv.Itoa(id),
		Body:       strings.NewReader(b.String()),
		Refresh:    "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%d", res.Status(), id)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
			return err
		}
		// Print the response status and indexed document version.
		log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
	}

	log.Println(strings.Repeat("-", 37))
	return nil
}

func (c *Client) searchDocuments(index string, keyword string) error {
	var r map[string]interface{}
	var buf bytes.Buffer
	// Create query struct.
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"title": "test",
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
		return err
	}

	// Perform the search request.
	res, err := c.es.Search(
		c.es.Search.WithContext(context.Background()),
		c.es.Search.WithIndex("test"),
		c.es.Search.WithBody(&buf),
		c.es.Search.WithTrackTotalHits(true),
		c.es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
			return err
		}
		// Print the response status and error information.
		log.Fatalf("[%s] %s: %s",
			res.Status(),
			e["error"].(map[string]interface{})["type"],
			e["error"].(map[string]interface{})["reason"],
		)
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
		return err
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	}

	log.Println(strings.Repeat("=", 37))
	return nil
}

func main() {
	log.SetFlags(0)
	c, err := newClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// 1. Get cluster info.
	c.getInfo()

	// 2. Index document.
	c.indexDocument("test", "test is good.", 3)

	// 3. Search document.
	c.searchDocuments("test", "test")
}
