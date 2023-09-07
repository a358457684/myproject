package elasticsearch

import (
	"common/config"
	"common/log"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
	"io/ioutil"
	"net/http"
)

var client *elasticsearch.Client

type EsData struct {
	Hits Hits `json:"hits"`
}

type Hits struct {
	Total Total    `json:"total"`
	Hits  []Source `json:"hits"`
}

type Total struct {
	Value int64 `json:"value"`
}

type Source struct {
	Source interface{} `json:"_source"`
}

func initElasticsearch(options *config.ElasticOptions) (*elasticsearch.Client, error) {
	log.Infof("开始初始化Elasticsearch=========")
	var err error
	client, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: options.Hosts,
		Username:  options.Username,
		Password:  options.Password,
	})
	return client, err
}

func parseEsData(res *esapi.Response, err error, query string, indices []string) Hits {
	if err != nil {
		log.WithError(err).Errorf("获取索引：%v的文档失败, 查询:%s, 返回信息：%+v", indices, query, res)
		return Hits{}
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusMultipleChoices {
		log.Errorf("获取索引：%v的文档失败, 查询:%s, 返回信息：%+v", indices, query, res)
		return Hits{}
	}
	log.Infof("获取索引：%v的文档成功, 查询:%s", indices, query)
	jsonData, _ := ioutil.ReadAll(res.Body)
	var esData EsData
	_ = json.Unmarshal(jsonData, &esData)
	return esData.Hits
}
