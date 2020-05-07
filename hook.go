package main

import (
    "fmt"
    "github.com/sirupsen/logrus"
    "io"
    "log"
    "os"
    "path/filepath"
    "reflect"
    "sync"
)



// We are logging to file, strip colors to make the output more readable.
var defaultFormatter = &Formatter{}

// PathMap is map for mapping a log level to a file's path.
// Multiple levels may share a file, but multiple files may not be used for one level.
type PathMap map[logrus.Level]string

// WriterMap is map for mapping a log level to an io.Writer.
// Multiple levels may share a writer, but multiple writers may not be used for one level.
type WriterMap map[logrus.Level]io.Writer

// RpcHook is a hook to handle writing to local log files.
type RpcHook struct {
    paths     PathMap
    writers   WriterMap
    levels    []logrus.Level
    lock      *sync.Mutex
    formatter logrus.Formatter
    currentindex int   //current file index
    maxindex     int   //max file index
    maxfilesize  int   //max file size
    defaultPath      string
    defaultWriter    io.Writer
    hasDefaultPath   bool
    hasDefaultWriter bool
}

// NewHook returns new LFS hook.
// Output can be a string, io.Writer, WriterMap or PathMap.
// If using io.Writer or WriterMap, user is responsible for closing the used io.Writer.
func NewHook(output interface{}, formatter logrus.Formatter) *RpcHook {
    hook := &RpcHook{
        lock: new(sync.Mutex),
    }
    
    hook.SetFormatter(formatter)
    
    switch output.(type) {
    case string:
        hook.SetDefaultPath(output.(string))
        break
    case io.Writer:
        hook.SetDefaultWriter(output.(io.Writer))
        break
    case PathMap:
        hook.paths = output.(PathMap)
        for level := range output.(PathMap) {
            hook.levels = append(hook.levels, level)
        }
        break
    case WriterMap:
        hook.writers = output.(WriterMap)
        for level := range output.(WriterMap) {
            hook.levels = append(hook.levels, level)
        }
        break
    default:
        panic(fmt.Sprintf("unsupported level map type: %v", reflect.TypeOf(output)))
    }
    
    return hook
}

// SetFormatter sets the format that will be used by hook.
// If using text formatter, this method will disable color output to make the log file more readable.
func (hook *RpcHook) SetFormatter(formatter logrus.Formatter) {
    hook.lock.Lock()
    defer hook.lock.Unlock()
    if formatter == nil {
        formatter = defaultFormatter
    } else {
        switch formatter.(type) {
        case *logrus.TextFormatter:
            textFormatter := formatter.(*logrus.TextFormatter)
            textFormatter.DisableColors = true
        }
    }
    
    hook.formatter = formatter
}

// SetDefaultPath sets default path for levels that don't have any defined output path.
func (hook *RpcHook) SetDefaultPath(defaultPath string) {
    hook.lock.Lock()
    defer hook.lock.Unlock()
    hook.defaultPath = defaultPath
    hook.hasDefaultPath = true
}

// SetDefaultWriter sets default writer for levels that don't have any defined writer.
func (hook *RpcHook) SetDefaultWriter(defaultWriter io.Writer) {
    hook.lock.Lock()
    defer hook.lock.Unlock()
    hook.defaultWriter = defaultWriter
    hook.hasDefaultWriter = true
}

// Fire writes the log file to defined path or using the defined writer.
// User who run this function needs write permissions to the file or directory if the file does not yet exist.
func (hook *RpcHook) Fire(entry *logrus.Entry) error {
    hook.lock.Lock()
    defer hook.lock.Unlock()
    if hook.writers != nil || hook.hasDefaultWriter {
        return hook.ioWrite(entry)
    } else if hook.paths != nil || hook.hasDefaultPath {
        return hook.fileWrite(entry)
    }
    
    return nil
}

// Write a log line to an io.Writer.
func (hook *RpcHook) ioWrite(entry *logrus.Entry) error {
    var (
        writer io.Writer
        msg    []byte
        err    error
        ok     bool
    )
    
    if writer, ok = hook.writers[entry.Level]; !ok {
        if hook.hasDefaultWriter {
            writer = hook.defaultWriter
        } else {
            return nil
        }
    }
    
    // use our formatter instead of entry.String()
    msg, err = hook.formatter.Format(entry)
    
    if err != nil {
        log.Println("failed to generate string for entry:", err)
        return err
    }
    _, err = writer.Write(msg)
    return err
}

// Write a log line directly to a file.
func (hook *RpcHook) fileWrite(entry *logrus.Entry) error {
    var (
        fd   *os.File
        path string
        msg  []byte
        err  error
        ok   bool
    )
    
    if path, ok = hook.paths[entry.Level]; !ok {
        if hook.hasDefaultPath {
            path = hook.defaultPath
        } else {
            return nil
        }
    }
    
    dir := filepath.Dir(path)
    os.MkdirAll(dir, os.ModePerm)
    
    fd, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
    if err != nil {
        log.Println("failed to open logfile:", path, err)
        return err
    }
    defer fd.Close()
    
    // use our formatter instead of entry.String()
    msg, err = hook.formatter.Format(entry)
    
    if err != nil {
        log.Println("failed to generate string for entry:", err)
        return err
    }
    fd.Write(msg)
    return nil
}

// Levels returns configured log levels.
func (hook *RpcHook) Levels() []logrus.Level {
    return logrus.AllLevels
}


