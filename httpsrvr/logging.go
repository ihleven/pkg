package httpsrvr


type logger interface{
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Fatal(err error, format string, args ...interface{})
}

