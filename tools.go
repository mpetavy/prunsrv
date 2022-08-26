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
)

func createLogFile(filename string) (*os.File, error) {
	dir := filepath.Dir(filename)
	if !fileExists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if checkError(err) {
			return nil, err
		}
	}

	if fileExists(filename) {
		fs, err := os.Stat(filename)
		if checkError(err) {
			return nil, err
		}

		if fs.Size() > 10000000 {
			debug(fmt.Sprintf("truncate log file %s ", filename))

			err = os.Remove(filename)
			if checkError(err) {
				return nil, err
			}
		}
	}

	var err error

	logf, err = os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
	if checkError(err) {
		return nil, err
	}

	return logf, nil
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
		log.Printf(fmt.Sprintf("%s %s", "ERROR", err.Error()))
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

	log.Printf(fmt.Sprintf("%s %s", "DEBUG", strings.Join(a, " ")))
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

func killProcess(pid int) error {
	p := findProcess(pid)
	if p == nil {
		return nil
	}

	err := p.Kill()
	if checkError(err) {
		return err
	}

	return nil
}

func findProcess(pid int) *os.Process {
	var b bool

	process, err := os.FindProcess(int(pid))
	if err != nil {
		process = nil
	} else {
		b = process.Signal(syscall.Signal(0)) == nil

		if !b {
			process = nil
		}
	}

	debug(fmt.Sprintf("findProcess %d: %v", pid, process != nil))

	return process
}

type mwriter struct {
	writers []io.Writer
}

func (mw *mwriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		w.Write(p)
	}

	return len(p), nil
}

func MWriter(writers ...io.Writer) io.Writer {
	allWriters := make([]io.Writer, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*mwriter); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}
	return &mwriter{allWriters}
}
