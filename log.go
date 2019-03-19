package icinga2


var log Logger = dummyLogger{}

type Logger interface {
	 Printf(format string, v ...interface{})
}

type dummyLogger struct {}

func (d dummyLogger)Printf(format string, v ...interface{}) {}
