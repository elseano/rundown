// +build debug

package util

import (
	"fmt"
)

func Debugf(s string, i ...interface{}) {
	fmt.Printf(s, i...)
}