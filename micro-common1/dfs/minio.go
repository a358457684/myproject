package dfs

import (
	"common/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client

func Init(options *config.DFSOptions) error {
	var err error
	Client, err = minio.New(options.Addr, &minio.Options{
		Creds:  credentials.NewStaticV4(options.Username, options.Password, ""),
		Secure: options.SSL,
	})
	return err
}
