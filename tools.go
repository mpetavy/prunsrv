package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unicode"
)

var (
	lastError string
	logf      *os.File
	mu        sync.Mutex
	logs      []string
)

func openLog() error {
	dir := configDir()
	if !fileExists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if isError(err) {
			return err
		}
	}

	var err error

	logf, err = os.OpenFile(filepath.Join(dir, title()+",log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if isError(err) {
		return err
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logf))

	return nil
}

func closeLog() {
	logf.Close()
}

func rerunElevated() error {
	debug("rerun elevated")

	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if isError(err) {
		return err
	}

	return nil
}

func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		debug("is running elevated: no")
		return false
	}

	debug("is running elevated: yes")

	return true
}

func checkAdmin() {
	if !isAdmin() {
		isError(rerunElevated())
	}
}

func hasFlag(flag string) bool {
	for _, arg := range os.Args {
		if arg == flag {
			return true
		}
	}

	return false
}

func isError(err error) bool {
	mu.Lock()
	defer mu.Unlock()

	if err == nil || err.Error() == lastError {
		return false
	}

	lastError = err.Error()

	s := fmt.Sprintf("%s %s", "ERROR", err.Error())
	logs = append(logs, s+"\n")

	log.Printf(s)

	return true
}

func debug(values ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	//if !hasFlag("--debug") {
	//	return
	//}

	var a []string

	for _, value := range values {
		a = append(a, fmt.Sprintf("%+v", reflect.ValueOf(value)))
	}

	s := fmt.Sprintf("%s %s", "DEBUG", strings.Join(a, " "))
	logs = append(logs, s+"\n")

	log.Printf(s)
}

func isWindowsOS() bool {
	b := runtime.GOOS == "windows"

	debug("isWindowsOs:", b)

	return b
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	var b bool

	if os.IsNotExist(err) || err != nil {
		b = false
	} else {
		b = true
	}

	debug("fileExists:", filename, b)

	return b
}

func javaExecutable() string {
	var s string
	if isWindowsOS() {
		s = "java"
	} else {
		s = "java"
	}

	debug("javaExecutable:", s)

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

	debug("title:", title)

	return title
}

func surroundWidth(strs []string, surround string, separator string) string {
	resultStrs := []string{}
	for _, str := range strs {
		resultStrs = append(resultStrs, fmt.Sprintf("\"%s\"", str))
	}

	result := strings.Join(resultStrs, separator)

	debug("surroundWidth:", result)

	return result
}
