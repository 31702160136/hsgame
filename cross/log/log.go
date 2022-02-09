package log

import (
	"cross/base"
	"cross/gtime"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	MODEL_LOG     byte = 1 //打印到日志
	MODEL_CONSOLE byte = 2 //打印到控制台
	MODEL_ALL     byte = 3 //打印到日志与控制台
	printFormat        = "%s %s [%s] %s"
	dateLen            = 10
	INFO               = "INFO"
	ERROR              = "ERROR"
)

var (
	logs      = make(map[string]*logFile)
	writerMux = sync.Mutex{}
	closeChan = make(chan byte, 0)
	logPath   = ""
)

type logFile struct {
	name  string
	model byte
	file  *os.File
	date  string
	sync.Mutex
	datas  []string
	suffix string
}

func NewLog(name string, model byte, suffix string) *logFile {
	return &logFile{
		name:   name,
		model:  model,
		datas:  make([]string, 0),
		suffix: suffix,
	}
}

func InitLog(path string) {
	logPath = path
	logs = map[string]*logFile{}
	switch runtime.GOOS {
	case "darwin":
		logs = map[string]*logFile{
			INFO:  NewLog(INFO, MODEL_ALL, "log"),
			ERROR: NewLog(ERROR, MODEL_ALL, "error"),
		}
	case "windows":
		logs = map[string]*logFile{
			INFO:  NewLog(INFO, MODEL_ALL, "log"),
			ERROR: NewLog(ERROR, MODEL_ALL, "error"),
		}
	default:
		logs = map[string]*logFile{
			INFO:  NewLog(INFO, MODEL_LOG, "log"),
			ERROR: NewLog(ERROR, MODEL_LOG, "error"),
		}
	}

	go func() {
		c := time.NewTimer(time.Millisecond * 200)
		isClose := false
		for {
			select {
			case <-c.C:
				println()
				c.Reset(time.Millisecond * 200)
			case <-closeChan:
				isClose = true
				break
			}
			if isClose {
				break
			}
		}
	}()
}

func Close() {
	closeChan <- 1
	println()
}

func println() {
	writerMux.Lock()
	for _, log := range logs {
		datas := log.GetDatas()
		if len(datas) == 0 {
			continue
		}

		if log.file == nil {
			log.refresh(gtime.Now().Format(gtime.DateFormat))
		}
		slash := ""
		if logPath[:len(logPath)-1] != "/" {
			slash = "/"
		}
		//如果文件不存在,则创建当天文件(防止人为删除文件)
		_, err := os.Stat(logPath + slash + log.date + "." + log.suffix)
		if os.IsNotExist(err) {
			log.refresh(gtime.Now().Format(log.date))
		}

		for _, data := range datas {
			//指定日期日志打印到指定日期的文件
			date := data[:dateLen]
			if data[:dateLen] != log.date {
				log.refresh(date)
			}

			switch log.model {
			case MODEL_LOG:
				log.file.WriteString(data + "\n")
			case MODEL_ALL:
				log.file.WriteString(data + "\n")
				fmt.Print(data + "\n")
			}
		}
	}
	writerMux.Unlock()
}

func (this *logFile) refresh(date string) {
	n := 10
	_ = createFile(logPath)
	slash := ""
	if logPath[:len(logPath)-1] != "/" {
		slash = "/"
	}
	this.date = date
	for i := 0; i < n; i++ {
		file, err := os.OpenFile(logPath+slash+date+"."+this.suffix, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err == nil {
			this.file = file
			break
		}
		if i+1 == n {
			fmt.Println(fmt.Sprintf(printFormat, gtime.Now().Format(gtime.DateTimeFormat), this.name, base.FileLine(3), "打开日志文件失败"))
		}
	}
}

func (this *logFile) writer(str string) {
	content := fmt.Sprintf(printFormat,
		gtime.Now().Format(gtime.DateTimeFormat),
		this.name,
		base.FileLine(3),
		str,
	)
	this.Lock()
	this.datas = append(this.datas, content)
	this.Unlock()
}

func (this *logFile) GetDatas() []string {
	this.Lock()
	datas := this.datas[:]
	this.datas = make([]string, 0)
	this.Unlock()
	return datas
}

func createFile(filePath string) error {
	if !folderIsExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}

func folderIsExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func Infof(format string, v ...interface{}) {
	logs[INFO].writer(fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	logs[INFO].writer(fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	logs[ERROR].writer(fmt.Sprintf(format, v...))
}
func Error(v ...interface{}) {
	logs[ERROR].writer(fmt.Sprint(v...))
}

func Fatalf(format string, v ...interface{}) {
	logs[ERROR].writer(fmt.Sprintf(format, v...))
	panic(fmt.Sprint(v...))
}
func Fatal(v ...interface{}) {
	logs[ERROR].writer(fmt.Sprint(v...))
	panic(fmt.Sprint(v...))
}
