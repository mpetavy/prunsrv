package main

import (
	"encoding/json"
	"fmt"
	"github.com/kardianos/service"
	"golang.org/x/sys/windows"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var logger service.Logger

type Pgosrv struct {
	DoTest        bool            `json:"-"`
	DoStart       bool            `json:"-"`
	DoStop        bool            `json:"-"`
	DoInstall     bool            `json:"-"`
	DoUninstall   bool            `json:"-"`
	DoUpdate      bool            `json:"-"`
	DoPrint       bool            `json:"-"`
	ServiceConfig service.Config  `json:"-"`
	Service       service.Service `json:"-"`
	Cmd           *exec.Cmd       `json:"-"`

	Name            string   `json:"Name"`
	Description     string   `json:"Description"`
	DisplayName     string   `json:"DisplayName"`
	StartPath       string   `json:"StartPath"`
	Startup         string   `json:"Startup"`
	JavaHome        string   `json:"JavaHome"`
	JvmOptions      []string `json:"JvmOptions"`
	Classpath       string   `json:"Classpath"`
	Jvm             string   `json:"Jvm"`
	JvmMx           string   `json:"JvmMx"`
	JvmMs           string   `json:"JvmMs"`
	JvmSs           string   `json:"JvmSs"`
	StartMode       string   `json:"StartMode"`
	StopMode        string   `json:"StopMode"`
	StartClass      string   `json:"StartClass"`
	StopClass       string   `json:"StopClass"`
	StartMethod     string   `json:"StartMethod"`
	StopMethod      string   `json:"StopMethod"`
	StopTimeout     int      `json:"StopTimeout"`
	LogPath         string   `json:"LogPath"`
	LogLevel        string   `json:"LogLevel"`
	LogPrefix       string   `json:"LogPrefix"`
	ServiceUser     string   `json:"ServiceUser"`
	ServicePassword string   `json:"ServicePassword"`
	PidFile         string   `json:"PidFile"`
}

func usage() {

}

func (p *Pgosrv) scanArgs() error {
	Debug("scanArgs")

	var err error

	argValue := func(arg string, i int) (string, int) {
		p := strings.Index(arg, "=")
		if p != -1 {
			s := strings.TrimSpace(arg[p+1:])

			return s, i
		}

		if len(arg) > 2 {
			p = strings.Index(arg[2:], "//")
			if p != -1 {
				return strings.TrimSpace(arg[p+4:]), i
			}
		}

		if i+1 < len(os.Args) {
			i++

			return strings.TrimSpace(os.Args[i]), i
		}

		panic(fmt.Errorf("Missing parameter to argument %s", arg))
	}

	for i := 1; i < len(os.Args); i++ {
		arg := strings.TrimSpace(os.Args[i])

		if strings.HasPrefix(arg, "//TS") {
			Debug("Action:", "testService")

			p.DoTest = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig()
			if Error(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//RS") || strings.HasPrefix(arg, "//ES") {
			Debug("Action:", "startService")

			p.DoStart = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig()
			if Error(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//SS") {
			Debug("Action:", "stopService")

			p.DoStop = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig()
			if Error(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//IS") {
			Debug("Action:", "installService")

			p.DoInstall = true

			p.Name, i = argValue(arg, i)
			p.Startup = "manual"
			p.Jvm = "auto"
			p.StartClass = "Main"
			p.StartMethod = "main"
			p.StopClass = "Main"
			p.StopMethod = "main"
			p.StopTimeout = 20
			p.LogPath = ""
			p.LogPrefix = title()
			p.LogLevel = "Info"
		}

		if strings.HasPrefix(arg, "//US") {
			Debug("Action:", "updateService")

			p.DoUpdate = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig()
			if Error(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//DS") {
			Debug("Action:", "deleteService")

			p.DoUninstall = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig()
			if Error(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//PS") {
			Debug("Action:", "printService")

			p.DoPrint = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig()
			if Error(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "--Description") {
			p.Description, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--DisplayName") {
			p.DisplayName, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StartPath") {
			p.StartPath, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--Startup") {
			p.Startup, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--JavaHome") {
			p.JavaHome, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--Classpath") {
			p.Classpath, i = argValue(arg, i)
		}

		if strings.Contains(arg, "JvmOptions") {
			var value string

			value, i = argValue(arg, i)

			values := strings.Split(value, ";")

			if strings.HasPrefix(arg, "++") {
				p.JvmOptions = append(p.JvmOptions, values...)
			} else {
				p.JvmOptions = values
			}
		}

		if strings.HasPrefix(arg, "--JvmMx") {
			p.JvmMx, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--JvmMs") {
			p.JvmMs, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--JvmSs") {
			p.JvmSs, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--Jvm=") || arg == "--Jvm" {
			p.Jvm, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StartMode") {
			p.StartMode, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopMode") {
			p.StopMode, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StartClass") {
			p.StartClass, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopClass") {
			p.StopClass, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StartMethod") {
			p.StartMethod, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopMethod") {
			p.StopMethod, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopTimeout") {
			var value string

			value, i = argValue(arg, i)

			p.StopTimeout, err = strconv.Atoi(value)
			if Error(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "--LogPath") {
			p.LogPath, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--LogPrefix") {
			p.LogPrefix, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--LogLevel") {
			p.LogLevel, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--ServiceUser") {
			p.ServiceUser, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--ServicePassword") {
			p.ServicePassword, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--PidFile") {
			p.PidFile, i = argValue(arg, i)
		}
	}

	return nil
}

func resolvEnvParameter(txt string) (string, error) {
	r, err := regexp.Compile(`\$\{.*?\}`)
	if Error(err) {
		return "", err
	}

	delta := 0
	pos := r.FindAllStringIndex(txt, -1)
	for _, p := range pos {
		p[0] += delta
		p[1] += delta
		env := txt[p[0]:p[1]]
		env = env[2 : len(env)-1]

		str := os.Getenv(env)
		if str != "" {
			txt = txt[:p[0]] + str + txt[p[1]:]

			delta += len(str) - (len(env) + 3)
		}
	}

	return txt, err
}

func (p *Pgosrv) exec(asStart bool) error {
	var args []string

	if p.JvmMx != "" {
		args = append(args, "-Xmx"+p.JvmMx)
	}
	if p.JvmMs != "" {
		args = append(args, "-Xms"+p.JvmMs)
	}
	if p.JvmSs != "" {
		args = append(args, "-Xss"+p.JvmSs)
	}

	for _, option := range p.JvmOptions {
		args = append(args, option)
	}

	if p.Classpath != "" {
		if p.StartClass != "" {
			args = append(args, "-cp")
		} else {
			args = append(args, "-jar")
		}

		args = append(args, p.Classpath)
	}

	if asStart {
		args = append(args, p.StartClass)
		args = append(args, p.StartMethod)
	} else {
		args = append(args, p.StopClass)
		args = append(args, p.StopMethod)
	}

	p.Cmd = &exec.Cmd{
		Path:   filepath.Join(p.JavaHome, "bin", javaExecutable()),
		Args:   args,
		Env:    nil,
		Dir:    p.StartPath,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	Debug("execCmd:", SurroundWidth(append([]string{p.Cmd.Path}, p.Cmd.Args...), "\"", " "))

	err := p.Cmd.Start()
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) Start(s service.Service) error {
	Debug("Start")

	err := p.exec(true)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) Stop(s service.Service) error {
	Debug("Stop")

	err := p.exec(false)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) printService() error {
	Debug("printService")

	args := []string{}
	args = append(args, title())
	args = append(args, fmt.Sprintf("//TS//%s", p.Name))
	args = append(args, fmt.Sprintf("%s=%s", "--Description", p.Description))
	args = append(args, fmt.Sprintf("%s=%s", "--DisplayName", p.DisplayName))
	args = append(args, fmt.Sprintf("%s=%s", "--StartPath", p.StartPath))
	args = append(args, fmt.Sprintf("%s=%s", "--Startup", p.Startup))
	args = append(args, fmt.Sprintf("%s=%s", "--JavaHome", p.JavaHome))
	args = append(args, fmt.Sprintf("%s=%s", "--JvmOptions", p.JvmOptions))
	args = append(args, fmt.Sprintf("%s=%s", "--Classpath", p.Classpath))
	args = append(args, fmt.Sprintf("%s=%s", "--Jvm", p.Jvm))
	args = append(args, fmt.Sprintf("%s=%s", "--JvmMx", p.JvmMx))
	args = append(args, fmt.Sprintf("%s=%s", "--JvmMs", p.JvmMs))
	args = append(args, fmt.Sprintf("%s=%s", "--JvmSs", p.JvmSs))
	args = append(args, fmt.Sprintf("%s=%s", "--StartMode", p.StopMode))
	args = append(args, fmt.Sprintf("%s=%s", "--StopMode", p.StopMode))
	args = append(args, fmt.Sprintf("%s=%s", "--StartClass", p.StartClass))
	args = append(args, fmt.Sprintf("%s=%s", "--StopClass", p.StopClass))
	args = append(args, fmt.Sprintf("%s=%s", "--StartMethod", p.StartMethod))
	args = append(args, fmt.Sprintf("%s=%s", "--StopMethod", p.StopMethod))
	args = append(args, fmt.Sprintf("%s=%s", "--StopTimeout", p.StopTimeout))
	args = append(args, fmt.Sprintf("%s=%s", "--LogPath", p.LogPath))
	args = append(args, fmt.Sprintf("%s=%s", "--LogLevel", p.LogLevel))
	args = append(args, fmt.Sprintf("%s=%s", "--LogPrefix", p.LogPrefix))
	args = append(args, fmt.Sprintf("%s=%s", "--ServiceUser", p.ServiceUser))
	args = append(args, fmt.Sprintf("%s=%s", "--ServicePassword", p.ServicePassword))
	args = append(args, fmt.Sprintf("%s=%s", "--PidFile", p.PidFile))

	fmt.Printf("%s\n", SurroundWidth(args, "\"", " "))

	return nil
}

func (p *Pgosrv) startService() error {
	Debug("startService")

	err := p.exec(true)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) stopService() error {
	Debug("stopService")

	err := p.exec(false)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) testService() error {
	Debug("testService")

	err := p.startService()
	if Error(err) {
		return err
	}

	var str string

	fmt.Scan(&str)

	//ctrlC := make(chan os.Signal, 1)
	//signal.Notify(ctrlC, os.Interrupt, syscall.SIGTERM)
	//
	//<-ctrlC

	err = p.stopService()
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) installService() error {
	Debug("installService")

	checkAdmin()

	err := p.saveConfig()
	if Error(err) {
		return err
	}

	err = p.loadConfig()
	if Error(err) {
		return err
	}

	p.Service, err = service.New(p, &p.ServiceConfig)
	if Error(err) {
		return err
	}

	service.Control(p.Service, "uninstall")

	err = service.Control(p.Service, "install")
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) updateService() error {
	Debug("installService")

	checkAdmin()

	err := p.saveConfig()
	if Error(err) {
		return err
	}

	p.Service, err = service.New(p, &p.ServiceConfig)
	if Error(err) {
		return err
	}

	service.Control(p.Service, "uninstall")

	err = service.Control(p.Service, "install")
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) uninstallService() error {
	Debug("uninstallService")

	checkAdmin()

	err := p.deleteConfig()
	Error(err)

	p.Service, err = service.New(p, &p.ServiceConfig)
	Error(err)

	err = service.Control(p.Service, "uninstall")
	Error(err)

	return nil
}

func configDir() string {
	var configDir string

	if isWindowsOS() {
		configDir = os.Getenv("ProgramData")
	} else {
		configDir = filepath.Join(string(filepath.Separator), "etc")
	}

	configDir = filepath.Join(configDir, title())

	Debug("configDir:", configDir)

	return configDir
}

func (p *Pgosrv) configFilename(dir string, extension string) string {
	filename := filepath.Join(dir, p.Name+extension)

	Debug("configFilename:", filename)

	return filename
}

func (p *Pgosrv) saveConfig() error {
	Debug("saveConfig")

	path := p.configFilename(configDir(), ".json")

	if !fileExists_(path) {
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if Error(err) {
			return err
		}
	}

	ba, err := json.MarshalIndent(p, "", "  ")
	if Error(err) {
		return err
	}

	err = ioutil.WriteFile(path, ba, os.ModePerm)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) loadConfig() error {
	Debug("loadConfig")

	path := p.configFilename(configDir(), ".json")

	if !fileExists_(path) {
		return fmt.Errorf("configuration is not available/readable: %s", path)
	}

	ba, err := ioutil.ReadFile(path)
	if Error(err) {
		return err
	}

	err = json.Unmarshal(ba, p)
	if Error(err) {
		return err
	}

	p.ServiceConfig.Name = p.Name
	p.ServiceConfig.Arguments = []string{fmt.Sprintf("//RS//%s", p.Name)}
	p.ServiceConfig.Description = p.Description
	p.ServiceConfig.DisplayName = p.DisplayName
	if p.ServiceConfig.DisplayName == "" {
		p.ServiceConfig.DisplayName = p.ServiceConfig.Name
	}
	p.ServiceConfig.UserName = p.ServiceUser
	if p.ServicePassword != "" {
		option := service.KeyValue{}
		option["Password"] = p.ServicePassword

		p.ServiceConfig.Option = option
	}

	return nil
}

func (p *Pgosrv) deleteConfig() error {
	Debug("deleteConfig")

	path := p.configFilename(configDir(), ".json")

	if !fileExists_(path) {
		return fmt.Errorf("configuration is not available or readable: %s", path)
	}

	err := os.RemoveAll(filepath.Dir(path))
	if Error(err) {
		return err
	}

	return nil
}

func rerunElevated() error {
	Debug("rerun elevated")

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
	if Error(err) {
		return err
	}

	return nil
}

func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		Debug("is running elevated: no")
		return false
	}

	Debug("is running elevated: yes")

	return true
}

func checkAdmin() {
	if !isAdmin() {
		Error(rerunElevated())
	}
}

func run() error {
	if len(os.Args) < 2 {
		usage()
	}

	Debug("cmdline:", SurroundWidth(os.Args, "\"", " "))

	dir := configDir()
	if !fileExists_(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if Error(err) {
			return err
		}
	}

	p := &Pgosrv{}

	err := p.scanArgs()
	if Error(err) {
		return err
	}

	if p.Name == "" {
		return fmt.Errorf("missing service name")
	}

	Debug("Service:", p.Name)

	if p.LogPath != "" {
		if fileExists_(p.LogPath) {
			f, err := os.OpenFile(p.configFilename(p.LogPath, ".log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
			if Error(err) {
				return err
			}

			for _, l := range logs {
				f.WriteString(l)
			}

			defer f.Close()

			log.SetPrefix(p.LogPrefix + " ")
			log.SetOutput(io.MultiWriter(os.Stdout, f))
		}
	}

	switch {
	case p.DoTest:
		return p.testService()
	case p.DoStart:
		return p.startService()
	case p.DoStop:
		return p.stopService()
	case p.DoInstall:
		return p.installService()
	case p.DoUpdate:
		return p.updateService()
	case p.DoUninstall:
		return p.uninstallService()
	case p.DoPrint:
		return p.printService()
	default:
		return fmt.Errorf("unknown action: %s", os.Args[1])
	}

	return nil
}

func main() {
	Error(run())
}
