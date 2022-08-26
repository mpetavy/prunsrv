package main

import (
	"encoding/json"
	"fmt"
	"github.com/kardianos/service"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var logger service.Logger

type Prunsrv struct {
	DoTest        bool            `json:"-"`
	DoService     bool            `json:"-"`
	DoStart       bool            `json:"-"`
	DoStop        bool            `json:"-"`
	DoInstall     bool            `json:"-"`
	DoUninstall   bool            `json:"-"`
	DoUpdate      bool            `json:"-"`
	DoPrint       bool            `json:"-"`
	ServiceConfig service.Config  `json:"-"`
	Service       service.Service `json:"-"`
	StartCmd      *exec.Cmd       `json:"-"`
	StopCmd       *exec.Cmd       `json:"-"`
	Logf          *os.File        `json:"-"`

	DisplayName     string   `json:"DisplayName"`
	Description     string   `json:"Description"`
	StartPath       string   `json:"StartPath"`
	Startup         string   `json:"Startup"`
	JavaHome        string   `json:"JavaHome"`
	JvmOptions      []string `json:"JvmOptions"`
	Classpath       string   `json:"Classpath"`
	JvmMx           string   `json:"JvmMx"`
	JvmMs           string   `json:"JvmMs"`
	JvmSs           string   `json:"JvmSs"`
	StartClass      string   `json:"StartClass"`
	StopClass       string   `json:"StopClass"`
	StartMethod     string   `json:"StartMethod"`
	StopMethod      string   `json:"StopMethod"`
	StopTimeout     string   `json:"StopTimeout"`
	LogPath         string   `json:"LogPath"`
	LogLevel        string   `json:"LogLevel"`
	LogPrefix       string   `json:"LogPrefix"`
	ServiceUser     string   `json:"ServiceUser"`
	ServicePassword string   `json:"ServicePassword"`
	PidFile         string   `json:"PidFile"`
}

const (
	version = "1.0.5"
)

func banner() {
	debug("banner")

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "%s %s - %s\n", strings.ToUpper(title()), version, "A GO based alternative to Apache PRUNSRV")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Copyright: © %s %s\n", "2022", "mpetavy")
	fmt.Fprintf(os.Stderr, "Homepage:  %s\n", "https://github.com/mpetavy/prunsrv")
	fmt.Fprintf(os.Stderr, "License:   %s\n", "https://www.apache.org/licenses/LICENSE-2.0.html")
	fmt.Fprintf(os.Stderr, "\n")
}

func usage() {
	debug("usage")

	fmt.Fprintf(os.Stderr, "Just developed to get around some Apache PRUNSRV problems.\n")
	fmt.Fprintf(os.Stderr, "Could not wait on Apache issue solving...\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "You are welcome to use this solution on your own risk.\n")
	fmt.Fprintf(os.Stderr, "Most of the original parameters are working the same.\n")
	fmt.Fprintf(os.Stderr, "Parameters which are not known are simply ignored.\n")
	fmt.Fprintf(os.Stderr, "Documentation can be found on the Github project page.\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Please consider the original PRUNSRV which can be found at\n")
	fmt.Fprintf(os.Stderr, "https://commons.apache.org/proper/commons-daemon/procrun.html\n")
}

func (p *Prunsrv) scanArgs() error {
	debug("scanArgs")

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

		panic(fmt.Errorf("missing parameter to argument %s", arg))
	}

	for i := 1; i < len(os.Args); i++ {
		arg := strings.TrimSpace(os.Args[i])

		if strings.HasPrefix(arg, "//TS") {
			debug("Action:", "testService")

			p.DoTest = true

			p.DisplayName, i = argValue(arg, i)

			err := p.loadConfig(true)
			if checkError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//RS") {
			debug("Action:", "runService")

			p.DoService = true

			p.DisplayName, i = argValue(arg, i)

			err := p.loadConfig(true)
			if checkError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//ES") {
			debug("Action:", "startService")

			p.DoStart = true

			p.DisplayName, i = argValue(arg, i)

			err := p.loadConfig(true)
			if checkError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//SS") {
			debug("Action:", "stopService")

			p.DoStop = true

			p.DisplayName, i = argValue(arg, i)

			err := p.loadConfig(true)
			if checkError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//IS") {
			debug("Action:", "installService")

			p.DoInstall = true

			p.DisplayName, i = argValue(arg, i)
			p.Startup = "manual"
			p.StartClass = "Service"
			p.StartMethod = "start"
			p.StopClass = "Service"
			p.StopMethod = "stop"
			p.StopTimeout = "20"
			p.LogPath = ""
			p.LogPrefix = title()
			p.LogLevel = "info"

			err := p.loadConfig(false)
			if checkError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//US") {
			debug("Action:", "updateService")

			p.DoUpdate = true

			p.DisplayName, i = argValue(arg, i)

			err := p.loadConfig(true)
			if checkError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//DS") {
			debug("Action:", "deleteService")

			p.DoUninstall = true

			p.DisplayName, i = argValue(arg, i)

			err := p.loadConfig(true)
			if checkError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//PS") {
			debug("Action:", "printService")

			p.DoPrint = true

			p.DisplayName, i = argValue(arg, i)

			err := p.loadConfig(true)
			if checkError(err) {
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
			p.StopTimeout, i = argValue(arg, i)
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

	p.ServiceConfig.Name = p.DisplayName
	p.ServiceConfig.Arguments = []string{fmt.Sprintf("//RS//%s", p.DisplayName)}
	p.ServiceConfig.Description = p.Description
	p.ServiceConfig.DisplayName = p.DisplayName
	if p.ServiceConfig.DisplayName == "" {
		p.ServiceConfig.DisplayName = p.ServiceConfig.Name
	}
	p.ServiceConfig.UserName = p.ServiceUser

	options := service.KeyValue{}

	if p.ServicePassword != "" {
		options["Password"] = p.ServicePassword
	}

	options["StartType"] = "automatic"

	switch p.Startup {
	case "manual":
		options["StartType"] = "manual"
	case "delayed":
		options["DelayedAutoStart"] = "true"
	case "disabled":
		options["StartType"] = "disabled"
	}

	p.ServiceConfig.Option = options

	p.Service, err = service.New(p, &p.ServiceConfig)
	if checkError(err) {
		return err
	}

	return nil
}

func resolvEnvParameter(txt string) (string, error) {
	r, err := regexp.Compile(`\$\{.*?\}`)
	if checkError(err) {
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

func (p *Prunsrv) exec(asStart bool) (*exec.Cmd, error) {
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

	logfWriter := io.Discard
	if p.Logf != nil {
		logfWriter = p.Logf
	}

	cmd := &exec.Cmd{
		Path:   filepath.Join(p.JavaHome, "bin", javaExecutable()),
		Args:   args,
		Env:    nil,
		Dir:    p.StartPath,
		Stdout: MWriter(logfWriter, os.Stdout),
		Stderr: MWriter(logfWriter, os.Stderr),
	}

	debug("execCmd:", strings.Join(surroundWidth(append([]string{cmd.Path}, cmd.Args...), "\""), " "))

	err := cmd.Start()
	if checkError(err) {
		return nil, err
	}

	return cmd, nil
}

func (p *Prunsrv) Start(s service.Service) error {
	debug("Start")

	go func() {
		checkError(p.startService())
	}()

	return nil
}

func (p *Prunsrv) Stop(s service.Service) error {
	debug("Stop")

	timeout, err := strconv.Atoi(p.StopTimeout)
	if checkError(err) {
		timeout = 60 * 60
	}

	timeoutDuration := time.Duration(min(timeout, 60*60)) * time.Second

	timeoutCh := time.NewTimer(timeoutDuration)

	err = p.stopService()
	if checkError(err) {
		return err
	}

	stopped := false
	for !stopped {
		select {
		case <-time.After(time.Millisecond * 500):
			stopped = stopped || findProcess(p.StartCmd.Process.Pid) == nil
		case <-timeoutCh.C:
			checkError(fmt.Errorf("process %d did not stop within %v, will kill it", p.StartCmd.Process.Pid, timeoutDuration))
			stopped = true
		}
	}

	if !timeoutCh.Stop() {
		<-timeoutCh.C
	}

	checkError(killProcess(p.StartCmd.Process.Pid))
	checkError(killProcess(p.StopCmd.Process.Pid))

	return nil
}

func (p *Prunsrv) printService() error {
	debug("printService")

	args := []string{}
	args = append(args, title())
	args = append(args, fmt.Sprintf("//TS//%s", p.DisplayName))
	args = append(args, fmt.Sprintf("%s=%s", "--Description", p.Description))
	args = append(args, fmt.Sprintf("%s=%s", "--DisplayName", p.DisplayName))
	args = append(args, fmt.Sprintf("%s=%s", "--StartPath", p.StartPath))
	args = append(args, fmt.Sprintf("%s=%s", "--Startup", p.Startup))
	args = append(args, fmt.Sprintf("%s=%s", "--JavaHome", p.JavaHome))
	for i := 0; i < len(p.JvmOptions); i++ {
		var prefix string
		if i == 0 {
			prefix = "--"
		} else {
			prefix = "++"
		}
		args = append(args, fmt.Sprintf("%s%s=%s", prefix, "JvmOptions", p.JvmOptions[i]))
	}
	args = append(args, fmt.Sprintf("%s=%s", "--Classpath", p.Classpath))
	args = append(args, fmt.Sprintf("%s=%s", "--JvmMx", p.JvmMx))
	args = append(args, fmt.Sprintf("%s=%s", "--JvmMs", p.JvmMs))
	args = append(args, fmt.Sprintf("%s=%s", "--JvmSs", p.JvmSs))
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

	var argSep string
	var lineSep string

	if isWindowsOS() {
		argSep = "\""
		lineSep = "^"
	} else {
		argSep = "'"
		lineSep = "\\"
	}

	args = surroundWidth(args, argSep)

	for i := 0; i < len(args)-1; i++ {
		args[i] = fmt.Sprintf("%s %s\n", args[i], lineSep)
	}

	fmt.Printf("%s\n", strings.Join(args, "  "))

	return nil
}

func (p *Prunsrv) startService() error {
	debug("startService")

	if p.LogPath != "" && p.LogLevel == "debug" {
		var err error

		p.Logf, err = createLogFile(p.configFilename(p.LogPath, ".log"))
		if checkError(err) {
			return err
		}
	}

	var err error

	p.StartCmd, err = p.exec(true)
	if checkError(err) {
		return err
	}

	b, pidfilename := getFlag("--PidFile")
	if b {
		checkError(ioutil.WriteFile(pidfilename, []byte(strconv.Itoa(p.StartCmd.Process.Pid)), os.ModePerm))
	}

	return nil
}

func (p *Prunsrv) stopService() error {
	debug("stopService")

	var err error

	p.StopCmd, err = p.exec(false)
	if checkError(err) {
		return err
	}

	if p.Logf != nil {
		err = p.Logf.Close()
		if checkError(err) {
			return err
		}
	}

	b, pidfilename := getFlag("--PidFile")
	if b && fileExists(pidfilename) {
		checkError(os.Remove(pidfilename))
	}

	return nil
}

func (p *Prunsrv) testService() error {
	debug("testService")

	err := p.startService()
	if checkError(err) {
		return err
	}

	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGTERM)

	<-ctrlC

	err = p.stopService()
	if checkError(err) {
		return err
	}

	return nil
}

func (p *Prunsrv) installService() error {
	debug("installService")

	checkAdmin()

	err := p.saveConfig()
	if checkError(err) {
		return err
	}

	err = p.loadConfig(true)
	if checkError(err) {
		return err
	}

	service.Control(p.Service, "stop")
	service.Control(p.Service, "uninstall")

	err = service.Control(p.Service, "install")
	if checkError(err) {
		return err
	}

	return nil
}

func (p *Prunsrv) updateService() error {
	debug("installService")

	checkAdmin()

	err := p.saveConfig()
	if checkError(err) {
		return err
	}

	service.Control(p.Service, "uninstall")

	err = service.Control(p.Service, "install")
	if checkError(err) {
		return err
	}

	return nil
}

func (p *Prunsrv) uninstallService() error {
	debug("uninstallService")

	checkAdmin()

	err := p.deleteConfig()
	checkError(err)

	service.Control(p.Service, "stop")

	err = service.Control(p.Service, "uninstall")
	if checkError(err) {
		return err
	}

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

	debug("configDir:", configDir)

	return configDir
}

func (p *Prunsrv) configFilename(dir string, extension string) string {
	filename := filepath.Join(dir, p.DisplayName+extension)

	debug("configFilename:", filename)

	return filename
}

func (p *Prunsrv) saveConfig() error {
	debug("saveConfig")

	filename := p.configFilename(configDir(), ".json")

	if !fileExists(filename) {
		err := os.MkdirAll(filepath.Dir(filename), os.ModePerm)
		if checkError(err) {
			return err
		}
	}

	ba, err := json.MarshalIndent(p, "", "  ")
	if checkError(err) {
		return err
	}

	err = ioutil.WriteFile(filename, ba, os.ModePerm)
	if checkError(err) {
		return err
	}

	return nil
}

func (p *Prunsrv) loadConfig(mustExist bool) error {
	debug("loadConfig")

	path := p.configFilename(configDir(), ".json")
	exists := fileExists(path)

	if !exists {
		if mustExist {
			return fmt.Errorf("configuration is not available/readable: %s", path)
		} else {
			return nil
		}
	}

	ba, err := ioutil.ReadFile(path)
	if checkError(err) {
		return err
	}

	err = json.Unmarshal(ba, p)
	if checkError(err) {
		return err
	}

	return nil
}

func (p *Prunsrv) deleteConfig() error {
	debug("deleteConfig")

	path := p.configFilename(configDir(), ".json")

	if !fileExists(path) {
		return fmt.Errorf("configuration is not available or readable: %s", path)
	}

	err := os.RemoveAll(filepath.Dir(path))
	if checkError(err) {
		return err
	}

	return nil
}

func run() error {
	b, _ := getFlag("--debug")
	isDebug = !service.Interactive() || b

	banner()

	var err error

	logf, err = createLogFile(filepath.Join(configDir(), title()+".log"))
	if checkError(err) {
		return err
	}

	defer func() {
		logf.Close()
	}()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(MWriter(logf, os.Stderr))

	b, _ = getFlag("//?")
	if len(os.Args) < 2 || b {
		usage()

		return nil
	}

	debug("cmdline:", strings.Join(surroundWidth(os.Args, "\""), " "))

	p := &Prunsrv{}

	err = p.scanArgs()
	if checkError(err) {
		return err
	}

	if p.DisplayName == "" {
		return fmt.Errorf("missing service name")
	}

	debug("Service:", p.DisplayName)

	switch {
	case p.DoService:
		return p.Service.Run()
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
	if checkError(run()) {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
