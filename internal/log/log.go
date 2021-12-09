package log

import "log"

var Logger = log.Default()

func init() {
	//Logger.SetPrefix()
}

func Info(s string) {
	Logger.Print("info: " + s)
}

func Infof(s string, v ...interface{}) {
	Logger.Printf("info: "+s, v...)
}

func Warn(s string) {
	Logger.Print("warn: " + s)
}

func Warnf(s string, v ...interface{}) {
	Logger.Printf("warn: "+s, v...)
}

func Error(err error, s string) {
	Logger.Printf(" err: " + err.Error() + s)
}

func Errorf(err error, s string, v ...interface{}) {
	Logger.Printf(" err: "+err.Error()+s, v...)
}
