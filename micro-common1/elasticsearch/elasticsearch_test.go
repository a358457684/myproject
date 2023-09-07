package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestSearch(t *testing.T) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"robotId": "26608668f82e48cca23a7fa0dfe29d24",
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}
	res, _ := client.Search(
		client.Search.WithContext(context.Background()),
		client.Search.WithIndex("robot_electric-2020.11"),
		client.Search.WithBody(&buf),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
	)
	fmt.Println(res)
	jsonData, _ := ioutil.ReadAll(res.Body)
	var esData EsData
	_ = json.Unmarshal(jsonData, &esData)
	fmt.Println(esData.Hits.Hits)
	defer res.Body.Close()
}

func TestPageByBody(t *testing.T) {
	dsl := `{"query": {"bool": {"must": [%s]}}}`
	var mustQuery []string
	mustQuery = append(mustQuery, fmt.Sprintf(`{"term": {"robotId": "%s"}}`, "26608668f82e48cca23a7fa0dfe29d24"))
	index := fmt.Sprintf("%s-%s", "robot_electric", time.Now().Format("2006.01"))
	sources := PageByBody(context.Background(), fmt.Sprintf(dsl, strings.Join(mustQuery, ",")),
		1, 10, "lastUploadTime:desc", index)
	fmt.Println(sources)
}

func TestCreateIndex(t *testing.T) {
	var body strings.Builder
	body.Reset()
	body.WriteString(`{"foo" : "bar `)
	body.WriteString("007007")
	body.WriteString(`	" }`)
	res, err := client.Index(
		"test2",
		strings.NewReader(body.String()),
		client.Index.WithDocumentID("007007"),
		client.Index.WithRefresh("true"),
		client.Index.WithPretty(),
		client.Index.WithTimeout(100),
		client.Index.WithContext(context.Background()),
	)
	fmt.Println(res, err)
	if err != nil {
		t.Fatalf("Error getting the response: %s", err)
	}
	defer res.Body.Close()
}

func TestCreateDocument(t *testing.T) {
	res, err := client.Create(
		"test2",
		"1",
		strings.NewReader(`{
		  "user": "kimchy",
		  "post_date": "2009-11-15T14:12:12",
		  "message": "trying out Elasticsearch"
		}`),
		client.Create.WithPretty(),
	)
	fmt.Println(res, err)
	if err != nil {
		t.Fatalf("Error getting the response: %s", err)
	}
	defer res.Body.Close()
}
