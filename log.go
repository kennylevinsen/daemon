package daemon

import "fmt"

// Log, an abstracted logger.
var log Logger

// Logger interface describes a basic logging interface.
type Logger interface {
	Fatal(v ...interface{})
	Print(v ...interface{})
}

// Fatal logs a fatal message.
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Fatalf logs a fatal message using a formatting string.
func Fatalf(format string, v ...interface{}) {
	log.Fatal(fmt.Sprintf(format, v...))
}

// Fatalln logs a fatal message with a newline appended to the end.
func Fatalln(v ...interface{}) {
	log.Fatal(fmt.Sprint(v...) + "\n")
}

// Print logs a message.
func Print(v ...interface{}) {
	log.Print(v...)
}

// Printf logs a message using a formatting string.
func Printf(format string, v ...interface{}) {
	log.Print(fmt.Sprintf(format, v...))
}

// Println logs a message with a newline appended to the end.
func Println(v ...interface{}) {
	log.Print(fmt.Sprint(v...) + "\n")
}

// SetLogger sets the daemon global logger.
func SetLogger(l Logger) {
	log = l
}

// GetLogger retrieves the daemon global logger.
func GetLogger() Logger {
	return log
}
