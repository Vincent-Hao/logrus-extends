package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultSize = 100

var DefaultPath = os.Getenv("GOPATH")

//用于srpc/log日志包中日志收集
var stdlog = log.New(os.Stdout, "srpc/log: ", log.LstdFlags|log.Lshortfile)

type LogFile struct {
	lock     *sync.Mutex
	file     *os.File
	filepath string
	filesize int64
	maxsize  int64
	curindex int64
	//maxindex     int64
	backendname string
	servicename string
	curdate     string //format:YYYYMMDD
}

func NewLogFile() *LogFile {
	return &LogFile{lock: new(sync.Mutex),
		maxsize:  DefaultSize,
		filepath: DefaultPath,
		curindex: 0,
		curdate:  time.Now().Format(defaultDateFormat),
	}
}
func (logfile *LogFile) SetMaxSize(size int64) {
	logfile.lock.Lock()
	defer logfile.lock.Unlock()
	logfile.maxsize = size
}

func (logfile *LogFile) SetBackendName(backendname string) {
	logfile.lock.Lock()
	defer logfile.lock.Unlock()
	logfile.backendname = backendname
}
func (logfile *LogFile) SetServiceName(servicename string) {
	logfile.lock.Lock()
	defer logfile.lock.Unlock()
	logfile.servicename = servicename
}
func (logfile *LogFile) SetCurDate(curdate string) {
	logfile.lock.Lock()
	defer logfile.lock.Unlock()
	logfile.curdate = curdate
}
func (logfile *LogFile) SetFilePath(fpath string) {
	logfile.lock.Lock()
	defer logfile.lock.Unlock()
	//dir := filepath.Dir(fpath)
	dirinfo, err := os.Stat(fpath)
	if err == nil && !dirinfo.IsDir() {
		err = fmt.Errorf("%s already exists and not a directory", fpath)
	}
	if os.IsExist(err) {
		if err := os.MkdirAll(fpath, 0755); err != nil {
			err = fmt.Errorf(" create %s directory error: %s", fpath, err)
		}
	}
	logfile.filepath = fpath
}
func (logfile *LogFile) Close() error {
	if logfile.file != nil {
		err := logfile.file.Close()
		return err
	}
	return nil
}
func (logfile *LogFile) SetFile() {
	logfile.lock.Lock()
	defer logfile.lock.Unlock()
	for {
		file, err := openFile(*logfile)
		if err != nil {
			err = fmt.Errorf("SetFile open log file %s error: %s", logfile.filepath, err)
			panic(err)
		}
		logfile.file = file
		fInfo, _ := logfile.file.Stat()
		if fInfo.Size() >= int64(logfile.maxsize*1024*1024) {
			logfile.curindex++
			if err := logfile.file.Close(); err != nil {
				stdlog.Println("close file error: ", err)
			}
		} else {
			//重启服务，从最后更新日志文件追
			logfile.filesize = fInfo.Size()
			break
		}
	}
}

func (logfile *LogFile) Write(data []byte) (n int, e error) {
	if logfile.file == nil {
		logfile.SetFile()
	}
	date := time.Now().Format("20060102")
	if logfile.curdate != date {
		logfile.SetCurDate(date)
	}
	OldIndex := logfile.curindex
	//fInfo, _ := logfile.file.Stat()
	//if fInfo.Size() >= logfile.maxsize*1024*1024 {
	//	logfile.lock.Lock()
	//	//LogFile 加锁后重新判断其他协程是否已经修改
	//	fInfo, _ = logfile.file.Stat()
	//	if OldIndex == logfile.curindex && fInfo.Size() >= logfile.maxsize*1024*1024 {
	//		logfile.curindex++
	//		oldfile := logfile.file
	//		file,err := openFile(*logfile)
	//		if err != nil {
	//			err = errors.New("write file open log file error: " + err.Error())
	//			panic(err)
	//		}
	//		logfile.file = file
	//		if err := oldfile.Close();err !=nil{
	//			stdlog.Println("close file error: ",err)
	//		}
	//	}
	//	logfile.lock.Unlock()
	//}
	if logfile.filesize >= logfile.maxsize*1024*1024 {
		logfile.lock.Lock()
		//LogFile 加锁后重新判断其他协程是否已经修改
		fInfo, _ := logfile.file.Stat()
		if OldIndex == logfile.curindex && fInfo.Size() >= logfile.maxsize*1024*1024 {
			logfile.curindex++
			oldfile := logfile.file
			file, err := openFile(*logfile)
			if err != nil {
				err = errors.New("write file open log file error: " + err.Error())
				panic(err)
			}
			logfile.file = file
			logfile.filesize = 0
			if err := oldfile.Close(); err != nil {
				stdlog.Println("close file error: ", err)
			}
		}
		logfile.lock.Unlock()
	}
	atomic.AddInt64(&logfile.filesize,int64(len(data)))
	n, e = logfile.file.Write(data)
	return n, e
}

func openFile(logfile LogFile) (*os.File, error) {
	if logfile.filepath == "" || logfile.backendname == "" || logfile.servicename == "" || logfile.curdate == "" {
		return nil, fmt.Errorf("filename can't empty")
	}
	fpath := logfile.filepath + "\\" + logfile.backendname + "." + logfile.servicename + "." +
		logfile.curdate + "." + fmt.Sprintf("%06d", logfile.curindex)
	file, err := os.OpenFile(fpath, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0664)
	if err != nil {
		return nil, fmt.Errorf("write file open log file %s error: %s", fpath, err)
	}
	return file, nil
}
