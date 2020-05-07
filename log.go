package main

import (
    "github.com/rifflock/lfshook"
    "github.com/sirupsen/logrus"
    "os"
    "strings"
    "time"
)

var (
    Logger      = logrus.New()
    timeFormat  = "15:04:05.000000"
    dateFormat  = "20060102"
)



func InitLog(cfg *LogCfg) {
    if cfg == nil{
        panic("config.LogCfg is nil.")
    }
    
    //初始化log level
    level := logLevelforCfg(cfg.LogLevel)
    //初始化writer
    writer := NewLogFile()
    writer.SetMaxSize(cfg.MaxFileSize)
    writer.SetBackendName(cfg.BackendName)
    writer.SetServiceName(cfg.ServerName)
    writer.SetCurDate(time.Now().Format("20060102"))
    fpath := os.Getenv("GOPATH")
    writer.SetFilePath(fpath)
    
    //初始化Formatter
    field := logfieldtoFormatMap(cfg.LogField)
    
    //初始化Logger变量
    Logger.SetReportCaller(true)
    Logger.SetLevel(level)
    Logger.SetFormatter(&Formatter{
        TimestampFormat: timeFormat,
        DateFormat: dateFormat,
        DisableDate:field[FieldKeyDate],
        DisableTimestamp:field[FieldKeyTime],
        DisableMicroSecond:field[FieldKeyMicroSecond],
        DisablePid:field[FieldKeyPid],
        DisableGoid:field[FieldKeyGoid],
    })
    
    //Logger.AddHook(newLfsHook())  //用hook处理文件多个输出流
    //Logger.SetOutput(ioutil.Discard)
    Logger.SetOutput(writer)//不同级别的日志输出到同一文件中

}
//close file pointer
func Close(){
    lg,ok := Logger.Out.(*LogFile)
    if ok {
        if lg.file != nil{
            err := lg.file.Close()
            if err != nil{
                stdlog.Println("log close file error: ",err)
            }
        }
    }
}
func logfieldtoFormatMap(logfield string) map[string]bool{
    field := make(map[string]bool)
    for i:=0; i<len(logfield); i++{
        if logfield[i] == '0'{
            if i == 0{
                field[FieldKeyDate] = true
            }
            if i == 1{
                field[FieldKeyTime] = true
            }
            if i == 2{
                field[FieldKeyMicroSecond] = true
            }
            if i == 3{
                field[FieldKeyPid] = true
            }
            if i == 4{
                field[FieldKeyGoid] = true
            }
        }
    }
    return field
}

//解析配置文件中loglevel
func logLevelforCfg(lv string) logrus.Level{
    var level logrus.Level
    
    switch strings.ToLower(lv) {
    case "panic":
        level = logrus.PanicLevel
    case "fatal":
        level = logrus.FatalLevel
    case "error":
        level = logrus.ErrorLevel
    case "warn":
        level = logrus.WarnLevel
    case "info":
        level = logrus.InfoLevel
    case "debug":
        level = logrus.DebugLevel
    case "trace":
        level = logrus.TraceLevel
    default:
        level = logrus.InfoLevel
    }
    return level
}

//用lfshook分流日志，不同级别的日志到不同的文件中
func newLfsHook() logrus.Hook {
    
    writer := NewLogFile()
    writer.SetMaxSize(2)
    writer.SetBackendName("sprc")
    writer.SetServiceName("fixbank.warning")
    writer.SetCurDate(time.Now().Format("20060102"))
    
    writer2 := NewLogFile()
    writer2.SetMaxSize(1)
    writer2.SetBackendName("sprc")
    writer2.SetServiceName("fixbank.info")
    writer2.SetCurDate(time.Now().Format("20060102"))
    lfsHook := lfshook.NewHook(lfshook.WriterMap{
        logrus.DebugLevel: writer,
        logrus.InfoLevel:  writer2,
        logrus.WarnLevel:  writer,
        logrus.ErrorLevel: writer,
        logrus.FatalLevel: writer,
        logrus.PanicLevel: writer,
    }, &Formatter{TimestampFormat: timeFormat, DateFormat: dateFormat,})
    
    return lfsHook
}

//如果使用自己封装接口，可以使用D结构传值
type D struct {
    Key string
    Value interface{}
}
// KV return a log kv for logging field.
func KV(key string, value interface{}) D {
    return D{
        Key:   key,
        Value: value,
    }
}
//服务中不需要银行号，使用此接口
func Trace(args ...interface{}){
    Logger.Trace(args...)
}
func Debug(args ...interface{}){
    Logger.Debug(args...)
}
func Info(args ...interface{}){
    Logger.Info(args...)
}
func Warn(args ...interface{}){
    Logger.Warn(args...)
}
func Error(args ...interface{}){
    Logger.Error(args...)
}
func Fatal(args ...interface{}){
    Logger.Fatal(args...)
}
func Panic(args ...interface{}){
    Logger.Panic(args...)
}
func WithFields(fields logrus.Fields) *logrus.Entry {
    if v,ok := fields[FieldKeyBankNo];ok{
        fields["fields."+FieldKeyBankNo] = v
    }
    return Logger.WithFields(fields)
}
