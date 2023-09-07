package elasticsearch

import (
	"common/config"
	"common/log"
	"context"
	"encoding/json"
	"errors"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"
)

func init() {
	var err error
	defer func() {
		if err != nil {
			log.WithError(err).Error("Elasticsearch初始化失败")
			panic(err)
		}
		log.Info("Elasticsearch初始化成功")
	}()
	if config.Data.Elastic == nil {
		err = errors.New("读取Elasticsearch配置失败")
		return
	}
	_, err = initElasticsearch(config.Data.Elastic)
}

// 根据DSL语句查询
func PageByBody(c context.Context, dsl string, pageIndex int, pageSize int, sort string, indices ...string) Hits {
	res, err := client.Search(
		client.Search.WithContext(c),
		client.Search.WithIndex(indices...),
		client.Search.WithBody(strings.NewReader(dsl)),
		client.Search.WithSort(sort),
		client.Search.WithFilterPath("hits.hits._source,hits.total"),
		client.Search.WithSize(pageSize),
		client.Search.WithFrom((pageIndex-1)*pageSize),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
		client.Search.WithIgnoreUnavailable(true),
	)
	return parseEsData(res, err, dsl, indices)
}

// 根据query查询
func PageByQuery(c context.Context, query string, pageIndex int, pageSize int, sort string, indices ...string) Hits {
	res, err := client.Search(
		client.Search.WithContext(c),
		client.Search.WithIndex(indices...),
		client.Search.WithQuery(query),
		client.Search.WithSort(sort),
		client.Search.WithFilterPath("hits.hits._source,hits.total"),
		client.Search.WithSize(pageSize),
		client.Search.WithFrom((pageIndex-1)*pageSize),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
		client.Search.WithIgnoreUnavailable(true),
	)
	return parseEsData(res, err, query, indices)
}

// 创建文档，自动创建索引
func CreateDocument(index, id string, data interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.WithError(err).Errorf("索引：%s, 添加文档失败, json转换失败", index)
		return err
	}
	if "" == id {
		id = strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	}
	res, err := client.Create(
		index,
		id,
		strings.NewReader(string(jsonBytes)),
		client.Create.WithPretty(),
		client.Create.WithDocumentType("default"),
	)
	if err != nil {
		log.WithError(err).Errorf("索引：%s, 添加文档失败, 返回信息：%+v", index, res)
		return errors.New("添加文档失败")
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusMultipleChoices {
		log.Errorf("索引：%s, 添加文档失败, 返回信息：%+v", index, res)
		return errors.New("添加文档失败")
	}
	return nil
}

/**
根据id修改文档
dsl 代表封装的需要改动的字段
如：{"doc": { "statusEndTime":"1608273482","timeConsume":"0.9"}}
*/
func UpdateDocument(index, id, dsl string) error {
	res, err := client.Update(
		index,
		id,
		strings.NewReader(dsl),
		client.Update.WithPretty(),
		client.Update.WithDocumentType("default"),
	)
	if err != nil {
		log.WithError(err).Errorf("索引：%s, 修改文档失败, 返回信息：%+v", index, res)
		return errors.New("修改文档失败")
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusMultipleChoices {
		log.Errorf("索引：%s, 修改文档失败, 返回信息：%+v", index, res)
		return errors.New("修改文档失败")
	}
	return nil
}

// 判断索引是否存在
func ExistIndices(indices ...string) bool {
	res, err := client.Indices.Exists(indices)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode < http.StatusMultipleChoices
}
