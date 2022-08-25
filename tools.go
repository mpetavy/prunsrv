package main

import (
	"fmt"
	"golang.org/x/exp/constraints"
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
	isDebug   bool
	logs      []string
)

func openLog() error {
	dir := configDir()
	if !fileExists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if checkError(err) {
			return err
		}
	}

	var err error

	logf, err = os.OpenFile(filepath.Join(dir, title()+",log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if checkError(err) {
		return err
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
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
	if checkError(err) {
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
		checkError(rerunElevated())
	}
}

func getFlag(flag string) (bool, string) {
	for i := 0; i < len(os.Args); i++ {
		arg := strings.TrimSpace(os.Args[i])

		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			p := strings.Index(arg, "=")
			if p != -1 {
				arg = strings.TrimSpace(arg[p+1:])

				return true, arg
			}

			if i+1 < len(os.Args) {
				return true, os.Args[i+1]
			}

			return true, ""
		}
	}

	return false, ""
}

func checkError(err error) bool {
	mu.Lock()
	defer mu.Unlock()

	if err == nil || err.Error() == lastError {
		return err != nil
	}

	lastError = err.Error()

	if isDebug {
		s := fmt.Sprintf("%s %s", "ERROR", err.Error())
		logs = append(logs, s+"\n")

		log.Printf(s)
	} else {
		fmt.Fprint(os.Stderr, err.Error()+"\n")
	}

	return true
}

func debug(values ...interface{}) {
	if !isDebug {
		return
	}

	mu.Lock()
	defer mu.Unlock()

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

func surroundWidth(strs []string, surround string) []string {
	resultStrs := []string{}
	for _, str := range strs {
		resultStrs = append(resultStrs, fmt.Sprintf("\"%s\"", str))
	}

	debug("surroundWidth:", resultStrs)

	return resultStrs
}

func max[T constraints.Ordered](v0 T, v1 T) T {
	if v0 > v1 {
		return v0
	}

	return v1
}

func min[T constraints.Ordered](v0 T, v1 T) T {
	if v0 < v1 {
		return v0
	}

	return v1
}

func killPid(pid int) error {
	p, err := os.FindProcess(1)
	if err == nil {
		return nil
	}

	err = p.Kill()
	if checkError(err) {
		return err
	}

	return nil
}
