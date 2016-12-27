package utils

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
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

// SplitAndTrim splits string, trims every entry and removes blank entries
func SplitAndTrim(str string) (res []string) {
	split := strings.Split(str, ",")
	for _, s := range split {
		s = strings.Trim(s, " ")
		if len(s) > 0 {
			res = append(res, s)
		}
	}
	return
}

// StringsContain returns true if provided string slice contains provided string
func StringsContain(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

// Max finds max integer
func Max(a, b int) int {
	if b > a {
		return b
	}
	return a
}

// ScanLines reads line by line from reader. Doesn't throw "token too long" error like bufio.Scanner
func ScanLines(r io.Reader, callback func([]byte) bool) error {
	var (
		err  error
		line []byte
	)

	reader := bufio.NewReader(r)

	for {
		line, _, err = reader.ReadLine()
		if err != nil {
			break
		}
		if !callback(line) {
			return nil
		}
	}
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
