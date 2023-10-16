package customlogger

import (
	"fmt"
	"io"
	"log"
	"time"
)

// define prefix map by color key
var prefixMap = map[string]string{
	"red":    "\033[31m",
	"green":  "\033[32m",
	"yellow": "\033[33m",
	"blue":   "\033[34m",
	"purple": "\033[35m",
	"cyan":   "\033[36m",
	"gray":   "\033[37m",
	"white":  "\033[97m",
}

// CustomLogger es un tipo que incorpora un logger de log y permite agregar un prefijo/sufijo.
type CustomLogger struct {
	Logger *log.Logger
	Prefix string
	Suffix string
}

func NewCustomLogger(color string, writer io.Writer) (*CustomLogger, error) {
	if _, ok := prefixMap[color]; !ok {
		return nil, fmt.Errorf("invalid color: %s", color)
	}

	return &CustomLogger{
		Logger: log.New(writer, "", 0),
		Prefix: prefixMap[color],
		Suffix: "\033[0m",
	}, nil
}

func (c *CustomLogger) Println(v ...interface{}) {
	msg := fmt.Sprint(v...)
	datetime := fmt.Sprintf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	msgWithPrefixAndSuffix := c.Prefix + datetime + msg + c.Suffix
	c.Logger.Println(msgWithPrefixAndSuffix)
}

func (c *CustomLogger) Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	datetime := fmt.Sprintf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	msgWithPrefixAndSuffix := c.Prefix + datetime + msg + c.Suffix
	c.Logger.Print(msgWithPrefixAndSuffix)
}
