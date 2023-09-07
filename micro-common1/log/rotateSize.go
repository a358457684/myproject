package log

import (
	"fmt"
	"github.com/suiyunonghen/DxCommonLib"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type rotateSizeWriter struct {
	splitSize  int
	logFile    string
	wChan      chan []byte
	bufferchan chan []byte
}

func (w *rotateSizeWriter) getBuffer(buflen int) (retbuf []byte) {
	var ok bool
	caplen := buflen
	if caplen < 512 {
		caplen = 512
	}
	select {
	case retbuf, ok = <-w.bufferchan:
		if !ok || cap(retbuf) < buflen {
			retbuf = make([]byte, buflen, caplen)
		}
	default:
		retbuf = make([]byte, buflen, caplen)
	}

	retbuf = retbuf[:buflen]
	return
}

func (w *rotateSizeWriter) reciveBuffer(buf []byte) bool {
	select {
	case w.bufferchan <- buf:
		return true
	case <-DxCommonLib.After(time.Second):
		//回收失败
		return false
	}
}

func (w *rotateSizeWriter) run() {
	wsize := 0
	logpath := path.Dir(w.logFile)
	if finfo, err := os.Stat(logpath); err != nil {
		os.MkdirAll(logpath, os.ModePerm)
	} else if !finfo.IsDir() {
		idx := 0
		for {
			logpath = logpath + strconv.Itoa(idx)
			finfo, err = os.Stat(logpath)
			if err != nil {
				os.MkdirAll(logpath, os.ModePerm)
				break
			} else {
				idx++
			}
		}
	}

	logfile := path.Base(w.logFile)
	basefile := logfile
	fileNameinfos := strings.FieldsFunc(logfile, func(r rune) bool {
		return r == '.'
	})
	ext := ""
	if len(fileNameinfos) > 1 {
		ext = "." + fileNameinfos[len(fileNameinfos)-1]
		basefile = strings.Join(fileNameinfos[:len(fileNameinfos)-1], "")
		logfile = logpath + "/" + basefile
	} else {
		logfile = logpath + "/" + logfile
	}
	basefile = strings.ToLower(basefile)
	fileDate := ""
	curfileName := ""

	//先找上一次的日志文件
	filepath.Walk(logpath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(strings.ToLower(info.Name()), basefile) {
			if fileDate == "" {
				fileDate = info.Name()[len(basefile):]
				curfileName = logpath + "/" + info.Name()
			} else {
				if strings.Compare(fileDate, info.Name()[len(basefile):]) < 0 {
					fileDate = info.Name()[len(basefile):]
					curfileName = logpath + "/" + info.Name()
				}
			}
		}
		return nil
	})

	if curfileName == "" {
		curfileName = logfile + time.Now().Format("2006-01-02_15_04_05") + ext
	} else {
		if finfo, err := os.Stat(curfileName); err == nil {
			wsize = int(finfo.Size())
			if wsize >= w.splitSize {
				curfileName = logfile + time.Now().Format("2006-01-02_15_04_05") + ext
			}
		}
	}
	if file, err := os.OpenFile(curfileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err == nil {
		file.Seek(0, io.SeekEnd) //移动到末尾
		for {
			select {
			case wbyte, ok := <-w.wChan:
				if !ok {
					file.Close()
					return
				}
				if wlen, err := file.Write(wbyte); err == nil {
					w.reciveBuffer(wbyte)
					wsize += wlen
					if wsize >= w.splitSize {
						file.Close()
						wsize = 0
						curfileName = logfile + time.Now().Format("2006-01-02_15_04_05") + ext
						if file, err = os.OpenFile(curfileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666);
							err != nil {
							fmt.Println("新建日志写入失败：", err.Error())
							return
						}
					}
				} else {
					w.reciveBuffer(wbyte)
				}
			}
		}
	} else {
		fmt.Println("启动日志写入失败：", err.Error())
	}
}

func (w *rotateSizeWriter) Write(p []byte) (n int, err error) {
	mp := w.getBuffer(len(p))
	copy(mp, p)
	DxCommonLib.PostFunc(func(data ...interface{}) {
		select {
		case w.wChan <- mp:
			return
		case <-DxCommonLib.After(time.Second):
			return
		}
	})
	return len(p), nil
}

func New(splitSize int, logfile string) *rotateSizeWriter {
	result := &rotateSizeWriter{wChan: make(chan []byte, 256), bufferchan: make(chan []byte, 256), splitSize: splitSize, logFile: logfile}
	go result.run()
	return result
}
