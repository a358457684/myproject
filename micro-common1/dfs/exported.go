package dfs

import (
	"common/config"
	"common/log"
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"io"
	"mime"
	"net/url"
	"path"
	"strings"
	"time"
)

func init() {
	if config.Data.DFS == nil {
		log.Error("读取Minio配置失败")
		return
	}
	err := Init(config.Data.DFS)
	if err != nil {
		log.WithError(err).Error("Minio初始化失败")
		panic(err)
	} else {
		log.Info("Minio初始化成功")
	}
}

// 判断存储桶是否存在
func BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return Client.BucketExists(ctx, bucketName)
}

// 创建存储桶
func MakeBucket(ctx context.Context, bucketName string) error {
	// 判断存储桶是否存在
	exists, err := BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		// 创建存储桶
		if err := Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
			return err
		}
	}
	return nil
}

// 列出所有存储桶
func ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return Client.ListBuckets(ctx)
}

// 上传文件
func PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error {
	bucketName := config.Data.DFS.Bucket
	info, err := Client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: mime.TypeByExtension(path.Ext(objectName))})
	if err != nil {
		// 判断是否是存储桶不存在
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) && minioErr.Code == "NoSuchBucket" {
			log.Warnf("存储桶[%s]不存在", bucketName)
			// 创建存储桶
			if err := MakeBucket(ctx, bucketName); err != nil {
				return err
			}
			log.Debugf("创建存储桶[%s]", bucketName)
			// 重新上传文件
			return PutObject(ctx, objectName, reader, objectSize)
		}
		return err
	}
	log.Infof("文件上传成功：%+v", info)
	return nil
}

// 下载文件
func GetObject(ctx context.Context, objectName string) (*minio.Object, error) {
	return Client.GetObject(ctx, config.Data.DFS.Bucket, objectName, minio.GetObjectOptions{})
}

// 从本地硬盘上传文件
func FPutObject(ctx context.Context, objectName, filePath string) error {
	bucketName := config.Data.DFS.Bucket
	ext := path.Ext(filePath)
	if path.Ext(objectName) == "" {
		objectName += ext
	}
	// 上传文件
	info, err := Client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: mime.TypeByExtension(ext)})
	if err != nil {
		// 判断是否是存储桶不存在
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr) && minioErr.Code == "NoSuchBucket" {
			log.Warnf("存储桶[%s]不存在", bucketName)
			// 创建存储桶
			if err := MakeBucket(ctx, bucketName); err != nil {
				return err
			}
			log.Debugf("创建存储桶[%s]", bucketName)
			// 重新上传文件
			return FPutObject(ctx, objectName, filePath)
		}
		return err
	}
	log.Infof("成功上传文件：%+v", info)
	return nil
}

// 将文件下载到硬盘
func FGetObject(ctx context.Context, objectName, filePath string) error {
	return Client.FGetObject(ctx, config.Data.DFS.Bucket, objectName, filePath, minio.GetObjectOptions{})
}

// 获取文件详细信息
func StatObject(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	return Client.StatObject(ctx, config.Data.DFS.Bucket, objectName, minio.GetObjectOptions{})
}

// 移除文件
func RemoveObject(ctx context.Context, objectName string) error {
	return Client.RemoveObject(ctx, config.Data.DFS.Bucket, objectName, minio.RemoveObjectOptions{})
}

// 拷贝文件
func CopyObject(ctx context.Context, destName, srcName string) error {
	info, err := Client.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: config.Data.DFS.Bucket,
		Object: destName,
	}, minio.CopySrcOptions{
		Bucket: config.Data.DFS.Bucket,
		Object: srcName,
	})
	if err != nil {
		return err
	}
	log.Infof("成功拷贝文件：%+v", info)
	return nil
}

// 移动文件
func MoveObject(ctx context.Context, destName, srcName string) error {
	if srcName == "" || destName == "" {
		return nil
	}
	if err := CopyObject(ctx, destName, srcName); err != nil {
		return err
	}
	if err := RemoveObject(ctx, srcName); err != nil {
		return err
	}
	log.Infof("成功拷贝文件：%s --> %s", srcName, destName)
	return nil
}

// 列出所有文件
func ListObjects(ctx context.Context, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return Client.ListObjects(ctx, config.Data.DFS.Bucket, opts)
}

// 创建临时上传链接，Body->binary方式上传
func PresignedPutObject(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
	objUrl, err := Client.PresignedPutObject(ctx, config.Data.DFS.Bucket, objectName, expiry)
	if err != nil {
		return objUrl, err
	}
	if config.Data.DFS.Domain != "" {
		domain := config.Data.DFS.Domain
		if !strings.HasPrefix(domain, "http") {
			domain = "http://" + domain
		}
		if p, err := url.Parse(domain); err != nil {
			return objUrl, err
		} else {
			objUrl.Scheme = p.Scheme
			objUrl.Host = p.Host
		}
	}
	return objUrl, err
}

// 创建临时下载链接
func PresignedGetObject(ctx context.Context, objectName string, expiry time.Duration) (*url.URL, error) {
	objUrl, err := Client.PresignedGetObject(ctx, config.Data.DFS.Bucket, objectName, expiry, url.Values{})
	if err != nil {
		return objUrl, err
	}
	if config.Data.DFS.Domain != "" {
		domain := config.Data.DFS.Domain
		if !strings.HasPrefix(domain, "http") {
			domain = "http://" + domain
		}
		if p, err := url.Parse(domain); err != nil {
			return objUrl, err
		} else {
			objUrl.Scheme = p.Scheme
			objUrl.Host = p.Host
		}
	}
	return objUrl, err
}

// 创建常规下载链接，无验证
func NormalGetObject(ctx context.Context, objectName string) (*url.URL, error) {
	objUrl, err := url.Parse(fmt.Sprintf("http://%s/%s/%s", config.Data.DFS.Addr, config.Data.DFS.Bucket, objectName))
	if err != nil {
		return objUrl, err
	}
	if config.Data.DFS.Domain != "" {
		domain := config.Data.DFS.Domain
		if !strings.HasPrefix(domain, "http") {
			domain = "http://" + domain
		}
		if p, err := url.Parse(domain); err != nil {
			return objUrl, err
		} else {
			objUrl.Scheme = p.Scheme
			objUrl.Host = p.Host
		}
	}
	return objUrl, err
}

// 上传文件并获取临时连接
func PutAndPresignedObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, expiry time.Duration) (*url.URL, error) {
	if err := PutObject(ctx, objectName, reader, objectSize); err != nil {
		return nil, err
	}
	return PresignedGetObject(ctx, objectName, expiry)
}
