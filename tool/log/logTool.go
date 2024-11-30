package log

import (
	"io"
	"log"
	"os"
)

func CreatLog() {
	createDir("log")
	f, err := os.OpenFile("./log/log.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		f.Close()
		log.Println("日志文件打开失败:", err)
		return
	}

	// 组合一下即可，os.Stdout代表标准输出流
	multiWriter := MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("-----------start----------")
}

// 判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createDir(_dir string) {
	exist, err := PathExists(_dir)

	if err != nil {
		log.Printf("get dir error![%v]\n", err)
		return
	}

	if exist {
		log.Printf("has dir![%v]\n", _dir)
	} else {
		log.Printf("no dir![%v]\n", _dir)
		// 创建文件夹
		err := os.Mkdir(_dir, os.ModePerm)
		if err != nil {
			log.Printf("mkdir failed![%v]\n", err)
		} else {
			log.Printf("mkdir success!\n")
		}
	}
}

// 参考 io.multiWriter ,修改后避免部分流关闭后报错
type multiWriter struct {
	writers []io.Writer
}

func (t *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			continue
		}
		if n != len(p) {
			err = io.ErrShortWrite
			continue
		}
	}
	return len(p), nil
}

var _ io.StringWriter = (*multiWriter)(nil)

func (t *multiWriter) WriteString(s string) (n int, err error) {
	var p []byte // lazily initialized if/when needed
	for _, w := range t.writers {
		if sw, ok := w.(io.StringWriter); ok {
			n, err = sw.WriteString(s)
		} else {
			if p == nil {
				p = []byte(s)
			}
			n, err = w.Write(p)
		}
		if err != nil {
			break
		}
		if n != len(s) {
			err = io.ErrShortWrite
			break
		}
	}
	return len(s), nil
}

func MultiWriter(writers ...io.Writer) io.Writer {
	allWriters := make([]io.Writer, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*multiWriter); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}
	return &multiWriter{allWriters}
}
