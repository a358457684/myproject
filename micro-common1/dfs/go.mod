module common/minio

go 1.18

require (
	common/config v0.0.0
	common/log v0.0.0-00010101000000-000000000000
	github.com/minio/minio-go/v7 v7.0.10
)

replace common/log => ../log

replace common/config => ../config
