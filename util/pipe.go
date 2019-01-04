package util

import "os"

func NewPipe() (read *os.File, write *os.File, err error) {
	read, write, err = os.Pipe()
	if err != nil {
		return
	}
	return
}
