package vm

import (
	"fmt"
)

type RunFunc func() (interface{}, error)

func Run(f RunFunc) (res interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("System error: %v", r)
		}
	}()
	return f()
}
