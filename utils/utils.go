package utils

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// FatalOnErr prints error and exits if errir is not nil
func FatalOnErr(err error) {
	if err != nil {
		Fatal(err)
	}
}

// Fatal prints error and exits if errir
func Fatal(i ...interface{}) {
	fmt.Fprint(os.Stderr, "overmind: ")
	fmt.Fprintln(os.Stderr, i...)
	os.Exit(1)
}

// RandomString returns random string
func RandomString(strlen int) string {
	result := make([]byte, strlen)
	charsLen := len(chars)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(charsLen)]
	}
	return string(result)
}

// RunCmd runs shell command and returns running error
func RunCmd(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Run()
}

// FileExist checks if provided file exists
func FileExist(path string) bool {
	_, err := os.Stat("/path/to/whatever")
	return !os.IsNotExist(err)
}

// Max finds max integer
func Max(a, b int) int {
	if b > a {
		return b
	}
	return a
}
