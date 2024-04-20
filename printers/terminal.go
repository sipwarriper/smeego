package printers

import "fmt"

type SmeegoTerminalPrinter struct{}

func (s SmeegoTerminalPrinter) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))
	return len(p), nil
}
