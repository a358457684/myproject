package util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path"

	"github.com/disintegration/imaging"
	"io"
	"mime/multipart"
	"os"
)

func SaveUploadfile(articlePicture multipart.File, oriname, filedir, staticDirPath string) (string, error) {
	MkPath(staticDirPath + filedir)
	name := CreateNormalUUID() + path.Ext(oriname)

	if Exists(staticDirPath + filedir + name) {
		RemoveDir(staticDirPath + filedir + name)
	}

	destFile, err := os.Create(staticDirPath + filedir + name)

	if err != nil {
		logrus.Errorf("Create failed: %s\n", err)
		return "", errors.New("创建文件失败")
	}
	defer destFile.Close()

	// 读取表单文件，写入保存文件
	_, err = io.Copy(destFile, articlePicture)
	if err != nil {
		logrus.Errorf("Write file failed: %s\n", err)
		return "", errors.New("文件写入失败")
	}
	return filedir + name, nil
}

func Deletefile(filepath string) {
	err := os.Remove(filepath)
	if err != nil {
		logrus.Errorf("delete file (%s)", err.Error())
	}
}

func MkPath(path string) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		logrus.Errorf("MkPath (%s)", err.Error())
	}
}

func Prokey(str1, str2 string) string {
	return fmt.Sprintf("%s_%s", str1, str2)
}

func SaveUploadThumbfile(articlePicture []byte, newFileName, filedir, staticDirPath string) (string, error) {

	subDir := "map/" + filedir + "/thumb/"
	realPath := staticDirPath + subDir

	MkPath(realPath)
	fileNameSuffix := path.Ext(newFileName)
	if fileNameSuffix == "" {
		fileNameSuffix = ".png"
	}
	name := CreateNormalUUID() + "_thumb" + fileNameSuffix

	freehandThumbPath := realPath + name

	if Exists(freehandThumbPath) {
		RemoveDir(freehandThumbPath)
	}

	buf := bytes.NewBuffer(articlePicture)
	image, err := imaging.Decode(buf)
	if err != nil {
		return "", err
	}

	image = imaging.Resize(image, 200, 150, imaging.Lanczos) //这里宽度有点问题 java 是按比例缩放
	err = imaging.Save(image, freehandThumbPath)

	if err != nil {
		return "", err
	}

	return subDir + name, nil
}

func Savefile(path string, name string, file multipart.File) {
	MkPath(path)

	destFile, err := os.Create(path + "/" + name)

	if err != nil {
		logrus.Errorf("Create failed: %s\n", err)
		return
	}
	defer destFile.Close()

	// 读取表单文件，写入保存文件
	_, err = io.Copy(destFile, file)
	if err != nil {
		logrus.Errorf("Write file failed: %s\n", err)
		return
	}
}

func Savefilebyte(file []byte, oriname, filedir, staticDirPath string) (string, error) {

	MkPath(staticDirPath + filedir)
	name := CreateNormalUUID() + path.Ext(oriname)

	filepath := staticDirPath + filedir + name

	err := ioutil.WriteFile(filepath, file, 0644)
	if err != nil {
		return "", err
	}

	return filedir + name, nil
}
