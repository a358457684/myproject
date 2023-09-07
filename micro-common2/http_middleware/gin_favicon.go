package http_middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func Favicon(path string) gin.HandlerFunc {
	path = filepath.FromSlash(path)
	if len(path) > 0 && !os.IsPathSeparator(path[0]) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = filepath.Join(wd, path)
	}

	info, err := os.Stat(path)
	if err != nil || info == nil || info.IsDir() {
		panic("Invalid favicon path: " + path)
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(file)

	return func(ctxt *gin.Context) {
		if ctxt.Request.RequestURI != "/favicon.ico" {
			ctxt.Next()
			return
		}
		if ctxt.Request.Method != "GET" && ctxt.Request.Method != "HEAD" {
			status := http.StatusOK
			if ctxt.Request.Method != "OPTIONS" {
				status = http.StatusMethodNotAllowed
			}
			ctxt.Header("Allow", "GET,HEAD,OPTIONS")
			ctxt.AbortWithStatus(status)
			ctxt.Abort()
			return
		}
		ctxt.Header("Content-Type", "image/x-icon")
		http.ServeContent(ctxt.Writer, ctxt.Request, "favicon.ico", info.ModTime(), reader)
		ctxt.Abort()
	}
}
