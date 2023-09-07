module elasticsearch

go 1.18

require (
	common/config v0.0.0
	common/log v0.0.0
	github.com/elastic/go-elasticsearch/v6 v6.8.10
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/satori/go.uuid v1.2.0
)

replace (
	common/config => ../config
	common/log => ../log
)
