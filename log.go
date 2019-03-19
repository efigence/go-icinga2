package icinga2

// Dummy package logger - to be replaced if you want to get partial deserialization errors
var log Logger = dummyLogger{}

type Logger interface {
	 Printf(format string, v ...interface{})
}

type dummyLogger struct {}

func (d dummyLogger)Printf(format string, v ...interface{}) {}
