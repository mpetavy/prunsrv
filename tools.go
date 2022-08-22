package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"unicode"
)

var (
	lastError string
)

func Error(err error) bool {
	if err == nil || err.Error() == lastError {
		return false
	}

	lastError = err.Error()
	log.Printf("%s %s", "ERROR", err.Error())

	return true
}

func Debug(values ...interface{}) {
	var a []string

	for _, value := range values {
		a = append(a, fmt.Sprintf("%+v", reflect.ValueOf(value)))
	}

	log.Printf("%s %s", "DEBUG", strings.Join(a, " "))
}

func isWindowsOS() bool {
	b := runtime.GOOS == "windows"

	Debug("isWindowsOs:", b)

	return b
}

func fileExists_(filename string) bool {
	_, err := os.Stat(filename)

	var b bool

	if os.IsNotExist(err) || err != nil {
		b = false
	} else {
		b = true
	}

	Debug("fileExists:", filename, b)

	return b
}

func javaExecutable() string {
	var s string
	if isWindowsOS() {
		s = "java"
	} else {
		s = "java"
	}

	Debug("javaExecutable:", s)

	return s
}

func title() string {
	path, err := os.Executable()
	if err != nil {
		path = os.Args[0]
	}

	path = filepath.Base(path)
	path = path[0:(len(path) - len(filepath.Ext(path)))]

	title := ""

	runes := []rune(path)
	for i := 0; i < len(runes); i++ {
		if string(runes[i]) == "-" {
			break
		}

		if unicode.IsLetter(runes[i]) {
			title = title + string(runes[i])
		}
	}

	Debug("title:", title)

	return title
}

func SurroundWidth(strs []string, surround string, separator string) string {
	resultStrs := []string{}
	for _, str := range strs {
		resultStrs = append(resultStrs, fmt.Sprintf("\"%s\"", str))
	}

	result := strings.Join(resultStrs, " ")

	Debug("surroundWidth:", result)

	return result
}
