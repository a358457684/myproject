package util

import (
	"encoding/base64"
	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//判断文件是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

//生成图片并保存到本地
func GenerateImage(imgStr, imgPath, imgName string) error {
	if !Exists(imgPath) {
		err := os.MkdirAll(imgPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	ddd, err := base64.StdEncoding.DecodeString(imgStr)
	if err != nil {
		return nil
	}
	err = ioutil.WriteFile(imgPath+imgName, ddd, 0666)
	return err
}

//将icon覆盖到图片上
func DrawImg(picDirectory, picName, coverPicPath string, x, y int) string {

	gimg, _ := gg.LoadImage(picDirectory + picName)
	coverimg, _ := gg.LoadImage(coverPicPath)
	ctx := gg.NewContextForImage(gimg)
	m := resize.Resize(uint(ctx.Width()/20), uint(ctx.Height()/20), coverimg, resize.Lanczos3)
	ctx.DrawImage(m, x, ctx.Height()-y)
	newPicPath := picDirectory + "robot_" + strings.Replace(uuid.NewV4().String(), "-", "", -1) + ".jpg"
	_ = gg.SaveJPG(newPicPath, ctx.Image(), 100)
	return newPicPath
}

//删除目录或文件
func RemoveDir(path string) bool {
	if IsDir(path) {
		err := removeContents(path)
		return err == nil
	} else {
		err := os.Remove(path)
		return err == nil
	}
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
