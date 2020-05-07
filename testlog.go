package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup
func printf(i int) {

	Logger.Info("printf: test log info: ", i)
	Logger.Trace("printf: test log trace: ", i)
	Logger.Debug("printf: test log debug: ", i)
	Logger.Warn("printf: test log warn: ", i)
	Logger.WithField("bank","6304").Error("printf: test log bankno is string 6304 ", i)
	Logger.WithField(FieldKeyBankNo,nil).Warn("printf: test log warn bankno is nil: ", i)
	Logger.WithField(FieldKeyBankNo,6304).Debug("printf: test log debug is int 6304: ", i)
	Logger.WithField(FieldKeyBankNo,"123456").Debug("printf: test log debug is string 123456: ", i)
	Logger.WithField(FieldKeyBankNo,654321).Debug("printf: test log debug is int 654321: ", i)
	
	//slog.WithFieldsBank(6401,logrus.Fields{"bankno":6304}).Error("printf: test log error: ", i)
	//slog.Logger.Fatal("printf: test log fatal: ", i)
	//slog.Logger.Panic("printf: test log Panic: ", i)
	wg.Done()
}

func main() {
	config,err := LoadYamlConfig()
	fmt.Println(err)
	InitLog(config)
	defer Close()
	var bank1 *int
	var bank2 *int64
	var bank3 *string
	bank4 := 1123456
	bank1 = &bank4
	bank5 := int64(2234567)
	bank2 = &bank5
	bank6 := "3345678"
	bank3 = &bank6
	start := time.Now()
	for i := 0; i < 30000; i++ {
		//slog.Info("main: test log info: ", i)
		//slog.WithFields(logrus.Fields{"main":i}).Info("main fields.")
		Logger.Info("logger info: ",i)
		Logger.WithField(FieldKeyBankNo,bank1).Debug("printf: test log debug is *int bank1: ", i)
		Logger.WithField(FieldKeyBankNo,bank2).Debug("printf: test log debug is *int64 bank2: ", i)
		Logger.WithField(FieldKeyBankNo,bank3).Debug("printf: test log debug is *string bank3: ", i)
	    wg.Add(1)
		go printf(i)
	}
	wg.Wait()
	fmt.Println("timeout: ",time.Now().Sub(start))
	//time.Sleep(1 * time.Second)
}
