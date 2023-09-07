package util

import (
	"bytes"
	"common/log"
	"encoding/base64"
	"github.com/disintegration/imaging"
	"io"
	"io/ioutil"
	"os"
)

func SaveBase64Image(data, directory string) string {
	if data == "" {
		log.Error("数据不能为空")
		return ""
	}
	fileName := CreateUUID() + ".png"
	directory = FormatPath(directory)
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		log.WithError(err).Error("文件夹创建失败")
		return ""
	}
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.WithError(err).Error("图片解析失败")
		return ""
	}
	err = ioutil.WriteFile(directory+fileName, bytes, 0644)
	if err != nil {
		log.WithError(err).Error("图片保存失败")
		return ""
	}
	return fileName
}

func ScaleImage(path string, scale float64, directory string) string {
	srcImage, err := imaging.Open(path)
	if err != nil {
		log.WithError(err).Error("图片打开失败")
		return ""
	}
	width := int(float64(srcImage.Bounds().Max.X) * scale)
	height := int((float64(srcImage.Bounds().Max.Y)) * scale)
	dstImage := imaging.Resize(srcImage, width, height, imaging.Lanczos)
	fileName := CreateUUID() + "_thumb.png"
	directory = FormatPath(directory)
	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		log.WithError(err).Error("文件夹创建失败")
		return ""
	}
	err = imaging.Save(dstImage, directory+fileName)
	if err != nil {
		log.WithError(err).Error("缩略图保存失败")
		return ""
	}
	return fileName
}

func ScaleImageIO(reader io.Reader, scale float64, format imaging.Format) (*bytes.Buffer, error) {
	srcImage, err := imaging.Decode(reader)
	if err != nil {
		log.WithError(err).Error("图片打开失败")
		return nil, err
	}
	width := int(float64(srcImage.Bounds().Max.X) * scale)
	height := int((float64(srcImage.Bounds().Max.Y)) * scale)
	dstImage := imaging.Resize(srcImage, width, height, imaging.Lanczos)
	result := &bytes.Buffer{}
	err = imaging.Encode(result, dstImage, format)
	return result, err
}
