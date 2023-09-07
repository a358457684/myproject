package dfs

import (
	"common/config"
	"common/log"
	"context"
	"errors"
	"github.com/minio/minio-go/v7"
	"io"
	"os"
	"testing"
	"time"
)

func TestListBuckets(t *testing.T) {
	log.Info(ListBuckets(context.Background()))
}

func TestBucketExists(t *testing.T) {
	log.Info(BucketExists(context.Background(), config.Data.DFS.Bucket))
}

func TestUploadFile(t *testing.T) {
	filePath := "./asdf.json"
	FPutObject(context.Background(), "qwer", filePath)
}

func TestFGetObject(t *testing.T) {
	FGetObject(context.Background(), "test2.json", "./test.json")
}

func TestGetObject(t *testing.T) {
	object, _ := GetObject(context.Background(), "qwer.json")
	localFile, _ := os.Create("./test.json")
	io.Copy(localFile, object)
}

func TestPutObject(t *testing.T) {
	file, _ := os.Open("config.yml")
	defer file.Close()
	fileStat, _ := file.Stat()
	PutObject(context.Background(), file.Name(), file, fileStat.Size())
}

func TestPresignedGetObject(t *testing.T) {
	url, _ := PresignedGetObject(context.Background(), "test2.json", time.Minute)
	log.Info(url)
}

func TestPresignedPutObject(t *testing.T) {
	log.Info(PresignedPutObject(context.Background(), "test3.png", time.Hour))
}

func TestRemoveObject(t *testing.T) {
	log.Info(RemoveObject(context.Background(), "test4.json"))
}

func TestListObjects(t *testing.T) {
	objects := ListObjects(context.Background(), minio.ListObjectsOptions{
		Prefix:    "test",
		Recursive: true,
	})
	for object := range objects {
		log.Info(object)
	}
}

func TestCopyObject(t *testing.T) {
	log.Info(CopyObject(context.Background(), "test3.json", "test2.json"))
}

func TestMoveObject(t *testing.T) {
	log.Info(MoveObject(context.Background(), "test5.json", "test4.json"))
}

type MyError struct {
	Code string
	error
}

func TestInit(t *testing.T) {
	file, _ := os.Open("config.yml")
	info, err := Client.PutObject(context.Background(), "asdf", "qwer", file, 12, minio.PutObjectOptions{})
	if err != nil {
		var myerr MyError
		if errors.As(err, &myerr) {
			if myerr.Error() == "NoSuchBucket" {
				log.Error("NoSuchBucket")
			}
		}

	}
	log.Info(info, err)
}

func TestStatObject(t *testing.T) {
	object, err := StatObject(context.TODO(), "public/sys_office.sql")
	log.Error(err)
	log.Infof("%+v", object)
}
