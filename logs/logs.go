package logs

import (
	"fmt"
	"os"
)

var OnExit = make([]func(), 0)

func AddOnExit(fn func()) {
	OnExit = append(OnExit, fn)
}

func Exit(n int) {
	for _, fn := range OnExit {
		fn()
	}
	os.Exit(n)
}

func Log(a ...any) {
	fmt.Println(a...)
}

func Logf(format string, a ...any) {
	fmt.Printf(format, a...)
}

func Error(a ...any) {
	fmt.Println("Error: ", fmt.Sprint(a...))
}

func Fatalf(format string, a ...any) {
	fmt.Println("Fatal Error: ", fmt.Sprintf(format, a...))
	Exit(1)
}

func Fatal(msg string, err error) {
	fmt.Println("Fatal Error:", msg, ": ", err)
	Exit(1)
}

func CheckFatal(msg string, err error) {
	if err != nil {
		Fatal(msg, err)
	}
}
