package main

import (
	"encoding/json"
	"fmt"
	"github.com/kardianos/service"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var logger service.Logger

type Program struct {
	DoTest        bool            `json:"-"`
	DoInstall     bool            `json:"-"`
	DoUninstall   bool            `json:"-"`
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

func (p *Program) scanArgs() error {
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
			p.Name, i = argValue(arg, i)
			p.DoTest = true
		}

		if strings.HasPrefix(arg, "//IS") || strings.HasPrefix(arg, "//US") {
			p.Name, i = argValue(arg, i)
			p.DoInstall = true
		}

		if strings.HasPrefix(arg, "//DS") {
			p.Name, i = argValue(arg, i)
			p.DoUninstall = true
		}

		if strings.HasPrefix(arg, "--Description=") {
			p.Description, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--DisplayName=") {
			p.DisplayName, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StartPath=") {
			p.StartPath, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--Startup=") {
			p.Startup, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--JavaHome=") {
			p.JavaHome, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--Classpath=") {
			p.Classpath, i = argValue(arg, i)
		}

		if strings.Contains(arg, "JvmOptions=") {
			var value string

			value, i = argValue(arg, i)

			if strings.HasPrefix(arg, "++") {
				p.JvmOptions = append(p.JvmOptions, value)
			} else {
				p.JvmOptions = []string{value}
			}
		}

		if strings.HasPrefix(arg, "--JvmMx=") {
			p.JvmMx, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--JvmMs=") {
			p.JvmMs, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--JvmSs=") {
			p.JvmSs, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--Jvm=") {
			p.Jvm, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StartMode=") {
			p.StartMode, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopMode=") {
			p.StopMode, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StartClass=") {
			p.StartClass, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopClass=") {
			p.StopClass, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StartMethod=") {
			p.StartMethod, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopMethod=") {
			p.StopMethod, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopMode=") {
			p.StopMode, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--StopTimeout=") {
			var value string

			value, i = argValue(arg, i)

			p.StopTimeout, err = strconv.Atoi(value)
			if Error(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "--LogPath=") {
			p.LogPath, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--LogPrefix=") {
			p.LogPrefix, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--LogLevel=") {
			p.LogLevel, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--ServiceUser=") {
			p.ServiceUser, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--ServicePassword=") {
			p.ServicePassword, i = argValue(arg, i)
		}

		if strings.HasPrefix(arg, "--PidFile=") {
			p.PidFile, i = argValue(arg, i)
		}
	}

	return nil
}

func (p *Program) exec(asStart bool) error {
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

	strs := []string{fmt.Sprintf("\"%s\"", p.Cmd.Path)}
	for _, arg := range p.Cmd.Args {
		strs = append(strs, fmt.Sprintf("\"%s\"", arg))
	}

	Debug("execCmd", strings.Join(strs, " "))

	err := p.Cmd.Start()
	if Error(err) {
		return err
	}

	return nil
}

func (p *Program) Start(s service.Service) error {
	Debug("Start")

	err := p.exec(true)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Program) Stop(s service.Service) error {
	Debug("Stop")

	err := p.exec(false)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Program) startService() error {
	Debug("startService")

	err := p.loadConfig()
	if Error(err) {
		return err
	}

	err = p.exec(true)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Program) stopService() error {
	Debug("stopService")

	err := p.loadConfig()
	if Error(err) {
		return err
	}

	err = p.exec(false)
	if Error(err) {
		return err
	}

	return nil
}

func (p *Program) testService() error {
	Debug("testService")

	if len(os.Args) > 2 {
		err := p.saveConfig()
		if Error(err) {
			return err
		}
	}

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

func (p *Program) installService() error {
	Debug("installService")

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

func (p *Program) uninstallService() error {
	Debug("uninstallService")

	err := p.loadConfig()
	if Error(err) {
		return err
	}

	err = p.deleteConfig()
	Error(err)

	p.Service, err = service.New(p, &p.ServiceConfig)
	Error(err)

	err = service.Control(p.Service, "uninstall")
	Error(err)

	return nil
}

func (p *Program) configFilename() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = string(filepath.Separator)
	}

	s := filepath.Join(configDir, title(), p.Name+".json")

	Debug("configFilename", s)

	return s
}

func (p *Program) saveConfig() error {
	Debug("saveConfig")

	path := p.configFilename()

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

func (p *Program) loadConfig() error {
	Debug("loadConfig")

	path := p.configFilename()

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
	p.ServiceConfig.UserName = p.ServiceUser
	if p.ServicePassword != "" {
		option := service.KeyValue{}
		option["Password"] = p.ServicePassword

		p.ServiceConfig.Option = option
	}

	return nil
}

func (p *Program) deleteConfig() error {
	Debug("deleteConfig")

	path := p.configFilename()

	if !fileExists_(path) {
		return fmt.Errorf("configuration is not available or readable: %s", path)
	}

	err := os.RemoveAll(filepath.Dir(path))
	if Error(err) {
		return err
	}

	return nil
}

//JavaHome         string
//JvmOptions       []string
//Classpath        string
//JvmMx            string
//StartMode        string
//StopMode         string
//StartClass       string
//StopClass        string
//StartMethod      string
//StopMethod       string
//StopTimeout      int
//LogPath          string
//LogLevel         string
//LogPrefix        string
//ServiceUser      string
//ServicePassword  string
//PidFile          string

func run() error {
	if len(os.Args) < 2 {
		usage()
	}

	p := &Program{
		DisplayName: "ServiceName",
		Startup:     "manual",
		JavaHome:    "%JAVA_HOME%",
		Jvm:         "auto",
		StartClass:  "Main",
		StartMethod: "main",
		StopClass:   "Main",
		StopMethod:  "main",
		StopTimeout: 0,
		LogPath:     "%SystemRoot%\\System32\\LogFiles\\Apache",
		LogPrefix:   "commons-daemon",
		LogLevel:    "Info",
	}

	err := p.scanArgs()
	if Error(err) {
		return err
	}

	switch {
	case p.DoTest:
		return p.testService()
	case p.DoInstall:
		return p.installService()
	case p.DoUninstall:
		return p.uninstallService()
	}

	return nil
}

func main() {
	run()
}
