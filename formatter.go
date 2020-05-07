package main

import (
	"bytes"
	"fmt"
	sequences "github.com/konsorten/go-windows-terminal-sequences"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)
const (
	defaultTimestampFormat = "15:04:05.000000"
	FieldKeyMsg            = "msg"
	FieldKeyLevel          = "level"
	FieldKeyTime           = "time"
	FieldKeyLogrusError    = "logrus_error"
	FieldKeyFunc           = "func"
	FieldKeyFile           = "file"
	FieldKeyDate           = "date"
	FieldKeyMicroSecond    = "microsecond"
	FieldKeyPid            = "pid"
	FieldKeyGoid           = "goid"
	defaultDateFormat      = "20060102"
    FieldKeyBankNo           = "bank"
)




type fieldKey string

// FieldMap allows customization of the key names for default fields.
type FieldMap map[fieldKey]string
func (f FieldMap) resolve(key fieldKey) string {
    if k, ok := f[key]; ok {
        return k
    }
    
    return string(key)
}
func prefixFieldClashes(data logrus.Fields, fieldMap FieldMap, reportCaller bool) {
	timeKey := fieldMap.resolve(FieldKeyTime)
	if t, ok := data[timeKey]; ok {
		data["fields."+timeKey] = t
		delete(data, timeKey)
	}
	
	msgKey := fieldMap.resolve(FieldKeyMsg)
	if m, ok := data[msgKey]; ok {
		data["fields."+msgKey] = m
		delete(data, msgKey)
	}
	
	levelKey := fieldMap.resolve(FieldKeyLevel)
	if l, ok := data[levelKey]; ok {
		data["fields."+levelKey] = l
		delete(data, levelKey)
	}
	
	logrusErrKey := fieldMap.resolve(FieldKeyLogrusError)
	if l, ok := data[logrusErrKey]; ok {
		data["fields."+logrusErrKey] = l
		delete(data, logrusErrKey)
	}
	
	// If reportCaller is not set, 'func' will not conflict.
	if reportCaller {
		funcKey := fieldMap.resolve(FieldKeyFunc)
		if l, ok := data[funcKey]; ok {
			data["fields."+funcKey] = l
		}
		fileKey := fieldMap.resolve(FieldKeyFile)
		if l, ok := data[fileKey]; ok {
			data["fields."+fileKey] = l
		}
	}
	dateKey := fieldMap.resolve(FieldKeyDate)
	if t, ok := data[dateKey]; ok {
		data["fields."+dateKey] = t
		delete(data, dateKey)
	}
	microsecondKey := fieldMap.resolve(FieldKeyMicroSecond)
	if t, ok := data[microsecondKey]; ok {
		data["fields."+microsecondKey] = t
		delete(data, microsecondKey)
	}
	pidKey := fieldMap.resolve(FieldKeyPid)
	if t, ok := data[pidKey]; ok {
		data["fields."+pidKey] = t
		delete(data, pidKey)
	}
	goidKey := fieldMap.resolve(FieldKeyGoid)
	if t, ok := data[goidKey]; ok {
		data["fields."+goidKey] = t
		delete(data, goidKey)
	}
	
	bankNo := fieldMap.resolve(FieldKeyBankNo)
	vl, ok := data[bankNo]
	if !ok || vl == nil{
		data[bankNo] = "0000"
	}else{
		t := reflect.TypeOf(vl)
		if t != nil {
			switch t.Kind() {
			case reflect.String:
				str := vl.(string)
				if len(str) >= 4{
					data[bankNo] = str[:4]
					break
				}
				bk,e := strconv.Atoi(str)
				if e != nil{
					data[bankNo] = "0000"
				}
				data[bankNo] = fmt.Sprintf("%04d",bk)
			case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
				vi := fmt.Sprintf("%04d",vl)
				data[bankNo] = vi[:4]
			case reflect.Ptr:
				switch t.Elem().Kind() {
				case reflect.String:
					str := vl.(*string)
					if len(*str) >= 4{
						data[bankNo] = (*str)[:4]
						break
					}
					bk,e := strconv.Atoi(*str)
					if e != nil{
						data[bankNo] = "0000"
					}
					data[bankNo] = fmt.Sprintf("%04d",bk)
				case reflect.Int:
					vi := vl.(*int)
					vs := fmt.Sprintf("%04d",*vi)
					data[bankNo] = vs[:4]
				case reflect.Int8:
					vi := vl.(*int8)
					vs := fmt.Sprintf("%04d",*vi)
					data[bankNo] = vs[:4]
				case reflect.Int16:
					vi := vl.(*int16)
					vs := fmt.Sprintf("%04d",*vi)
					data[bankNo] = vs[:4]
				case reflect.Int32:
					vi := vl.(*int32)
					vs := fmt.Sprintf("%04d",*vi)
					data[bankNo] = vs[:4]
				case reflect.Int64:
					vi := vl.(*int64)
					vs := fmt.Sprintf("%04d",*vi)
					data[bankNo] = vs[:4]
				default:
					data[bankNo] = "0000"
				}
			default:
				data[bankNo] = "0000"
			}
		}
	}
}
func checkIfTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		var mode uint32
		err := syscall.GetConsoleMode(syscall.Handle(v.Fd()), &mode)
		return err == nil
	default:
		return false
	}
}
func initTerminal(w io.Writer) {
	switch v := w.(type) {
	case *os.File:
		sequences.EnableVirtualTerminalProcessing(syscall.Handle(v.Fd()), true)
	}
}

//get goroutine id
func Goid() int {
	defer func()  {
		if err := recover(); err != nil {
			fmt.Println("panic recover:panic info:%v", err)        }
	}()
	
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}


// TextFormatter formats logs into text
type Formatter struct {
	
	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool
	
	// TimestampFormat to use for display when a full timestamp is printed
	TimestampFormat string
	DateFormat string
	
	//pid
	DisableDate bool
	DisablePid bool
	DisableGoid bool
	DisableMicroSecond bool
	
	// QuoteEmptyFields will wrap empty fields in quotes if true
	QuoteEmptyFields bool
	
	// Whether the logger's out is to a terminal
	isTerminal bool
	
	// FieldMap allows users to customize the names of keys for default fields.

	FieldMap FieldMap
	
	// CallerPrettyfier can be set by the user to modify the content
	// of the function and file keys in the data when ReportCaller is
	// activated. If any of the returned value is the empty string the
	// corresponding key will be removed from fields.
	CallerPrettyfier func(*runtime.Frame) (function string, file string)
	
	terminalInitOnce sync.Once
}


func (f *Formatter) init(entry *logrus.Entry) {
	if entry.Logger != nil {
		f.isTerminal = checkIfTerminal(entry.Logger.Out)
		
		if f.isTerminal {
			initTerminal(entry.Logger.Out)
		}
	}
}

// Format renders a single log entry
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields)
	for k, v := range entry.Data {
		data[k] = v
	}
	
	prefixFieldClashes(data, f.FieldMap, entry.HasCaller())
	
	
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	
	var funcVal, fileVal string
	
	fixedKeys := make([]string, 0, 8+len(data))
	if !f.DisableDate{
		fixedKeys = append(fixedKeys,f.FieldMap.resolve(FieldKeyDate))
	}
	if !f.DisableTimestamp {
		fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyTime))
	}
	if !f.DisableMicroSecond {
		fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyMicroSecond))
	}
	if !f.DisablePid {
		fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyPid))
	}
	if !f.DisableGoid {
		fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyGoid))
	}
	
	fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyLevel))
	
	
	//if entry.err != "" {
	//	fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyLogrusError))
	//}
	if entry.HasCaller() {
		//CallerPrettyfier是entry.caller处理函数接口，目前没有使用，直接在else处理
		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		} else {
			//Val := entry.Caller.Function
			funcStr := strings.Split(entry.Caller.Function,".")
			funcVal = funcStr[len(funcStr)-1]
			file := entry.Caller.File
			//只取文件上层目录
			count := 0
			for i := len(file) - 1; i>0 ; i--{
			    if file[i] == '/' {
					count++
					if count >=1 {
						file = file[i+1:]
						break
					}
				}
			}
			fileVal = fmt.Sprintf("%s:%d", file, entry.Caller.Line)
		}
		
		if funcVal != "" {
			//fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldBankNo))
			fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyFunc))
		}
		if fileVal != "" {
			fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyFile))
		}
	}
	if entry.Message != "" {
		fixedKeys = append(fixedKeys, f.FieldMap.resolve(FieldKeyMsg))
	}

	fixedKeys = append(fixedKeys, keys...)
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	
	f.terminalInitOnce.Do(func() { f.init(entry) })
	
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}
	dateFormat := f.DateFormat
	if dateFormat == "" {
		dateFormat = defaultDateFormat
	}
	var microsecond string
	var clock string
	if !f.DisableMicroSecond || !f.DisableTimestamp{
		time := entry.Time.Format(timestampFormat)
		str := strings.Split(time,".")
		microsecond = str[len(str)-1]
		clock = str[0]
	}
	for _, key := range fixedKeys {
		var value interface{}
		flag := false
		switch {
		case key == f.FieldMap.resolve(FieldKeyTime):
			value = clock
		case key == f.FieldMap.resolve(FieldKeyLevel):
			value,_ = LeveltoCupData(entry.Level)
		case key == f.FieldMap.resolve(FieldKeyMsg):
			value = entry.Message
		//case key == f.FieldMap.resolve(FieldKeyLogrusError):
		//	value = entry.err
		case key == f.FieldMap.resolve(FieldKeyFunc) && entry.HasCaller():
			value = fmt.Sprintf("%s    ", data[FieldKeyBankNo])+funcVal
		case key == f.FieldMap.resolve(FieldKeyBankNo):
			continue
		case key == f.FieldMap.resolve(FieldKeyFile) && entry.HasCaller():
			value = fileVal
		case key == f.FieldMap.resolve(FieldKeyDate):
			value = entry.Time.Format(dateFormat)
		case key == f.FieldMap.resolve(FieldKeyMicroSecond):
			value = microsecond
		case key == f.FieldMap.resolve(FieldKeyPid):
			value = os.Getpid()
			flag = true
		case key == f.FieldMap.resolve(FieldKeyGoid):
			value = Goid()
			flag = true
		default:
			value = data[key]
			flag = true
		}
		
		f.appendKeyValue(b, key, value,flag)
	}
	
	b.WriteByte('\n')
	return b.Bytes(), nil
}
/*
func (f *CupDataFormatter) needsQuoting(text string) bool {
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}
*/


func (f *Formatter) appendKeyValue(b *bytes.Buffer, key string, value interface{},flag bool) {
	/*if b.Len() > 0 {
		b.WriteByte(' ')
	}*/
	b.WriteByte('[')
	if flag {
		b.WriteString(key)
		b.WriteByte(' ')
		b.WriteByte('=')
		b.WriteByte(' ')
	}
	
	f.appendValue(b, value)
	b.WriteByte(']')
}

func (f *Formatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	
	//if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	//} else {
		//b.WriteString(fmt.Sprintf("%q", stringVal))
		//b.WriteString(fmt.Sprintf("%s", stringVal))
	//}
}

func LeveltoCupData(level logrus.Level) (string, error) {
	switch level {
	case logrus.TraceLevel:
		return "LOGTRC", nil
	case logrus.DebugLevel:
		return "LOGDBG", nil
	case logrus.InfoLevel:
		return "LOGINF", nil
	case logrus.WarnLevel:
		return "LOGWAN", nil
	case logrus.ErrorLevel:
		return "LOGERR", nil
	case logrus.FatalLevel:
		return "LOGFAT", nil
	case logrus.PanicLevel:
		return "LOGPAC", nil
	}
	
	return "", fmt.Errorf("not a valid logrus level %d", level)
}