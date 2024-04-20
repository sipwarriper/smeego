package smeego

import (
	"fmt"
	"io"
)

func printError(printer io.Writer, message string, err error) {
	if printer == nil {
		return
	}
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("%s: %v", message, err)
	} else {
		msg = message
	}
	printer.Write([]byte(msg))
}

func printDebug(printer io.Writer, message string) {
	if printer == nil {
		return
	}
	printer.Write([]byte(message))
}
