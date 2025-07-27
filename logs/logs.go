package logs

import (
	"fmt"
	"os"
)

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
	os.Exit(1)
}

func Fatal(msg string, err error) {
	fmt.Println("Fatal Error:", msg, ": ", err)
	os.Exit(1)
}

func CheckFatal(msg string, err error) {
	if err != nil {
		Fatal(msg, err)
	}
}
