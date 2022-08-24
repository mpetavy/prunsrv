package main

import (
	"encoding/json"
	"fmt"
	"github.com/kardianos/service"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var logger service.Logger

type Pgosrv struct {
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

const (
	version = "1.0.0"
)

func banner() {
	fmt.Printf("\n")
	fmt.Printf("%s %s %s\n", strings.ToUpper(title()), version, "A GO based alternative to Apache PRUNSRV")
	fmt.Printf("\n")
	fmt.Printf("Copyright: © %s %s\n", "2022", "mpetavy")
	fmt.Printf("Homepage:  %s\n", "https://github.com/mpetavy/prunsrv")
	fmt.Printf("License:   %s\n", "https://www.apache.org/licenses/LICENSE-2.0.html")
	fmt.Printf("\n")
}

func usage() {
	debug("usage")

	fmt.Println()
	fmt.Printf("%s - some alternative to Apache Commons...\n", strings.ToUpper(title()))
	fmt.Printf("Documentation: https://commons.apache.org/proper/commons-daemon/procrun.html\n")
}

func (p *Pgosrv) scanArgs() error {
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

			p.Name, i = argValue(arg, i)

			err := p.loadConfig(true)
			if isError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//RS") {
			debug("Action:", "runService")

			p.DoService = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig(true)
			if isError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//ES") {
			debug("Action:", "startService")

			p.DoStart = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig(true)
			if isError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//SS") {
			debug("Action:", "stopService")

			p.DoStop = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig(true)
			if isError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//IS") {
			debug("Action:", "installService")

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

			err := p.loadConfig(false)
			if isError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//US") {
			debug("Action:", "updateService")

			p.DoUpdate = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig(true)
			if isError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//DS") {
			debug("Action:", "deleteService")

			p.DoUninstall = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig(true)
			if isError(err) {
				return err
			}
		}

		if strings.HasPrefix(arg, "//PS") {
			debug("Action:", "printService")

			p.DoPrint = true

			p.Name, i = argValue(arg, i)

			err := p.loadConfig(true)
			if isError(err) {
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
			if isError(err) {
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

	p.Service, err = service.New(p, &p.ServiceConfig)
	if isError(err) {
		return err
	}

	return nil
}

func resolvEnvParameter(txt string) (string, error) {
	r, err := regexp.Compile(`\$\{.*?\}`)
	if isError(err) {
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

	debug("execCmd:", surroundWidth(append([]string{p.Cmd.Path}, p.Cmd.Args...), "\"", " "))

	err := p.Cmd.Start()
	if isError(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) Start(s service.Service) error {
	debug("Start")

	go func() {
		isError(p.exec(true))
	}()

	return nil
}

func (p *Pgosrv) Stop(s service.Service) error {
	debug("Stop")

	go func() {
		isError(p.exec(false))
	}()

	return nil
}

func (p *Pgosrv) printService() error {
	debug("printService")

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

	fmt.Printf("%s\n", surroundWidth(args, "\"", " "))

	return nil
}

func (p *Pgosrv) startService() error {
	debug("startService")

	err := p.exec(true)
	if isError(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) stopService() error {
	debug("stopService")

	err := p.exec(false)
	if isError(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) testService() error {
	debug("testService")

	err := p.startService()
	if isError(err) {
		return err
	}

	var str string

	fmt.Scan(&str)

	//ctrlC := make(chan os.Signal, 1)
	//signal.Notify(ctrlC, os.Interrupt, syscall.SIGTERM)
	//
	//<-ctrlC

	err = p.stopService()
	if isError(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) installService() error {
	debug("installService")

	checkAdmin()

	err := p.saveConfig()
	if isError(err) {
		return err
	}

	err = p.loadConfig(true)
	if isError(err) {
		return err
	}

	service.Control(p.Service, "uninstall")

	err = service.Control(p.Service, "install")
	if isError(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) updateService() error {
	debug("installService")

	checkAdmin()

	err := p.saveConfig()
	if isError(err) {
		return err
	}

	service.Control(p.Service, "uninstall")

	err = service.Control(p.Service, "install")
	if isError(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) uninstallService() error {
	debug("uninstallService")

	checkAdmin()

	err := p.deleteConfig()
	isError(err)

	err = service.Control(p.Service, "uninstall")
	isError(err)

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

func (p *Pgosrv) configFilename(dir string, extension string) string {
	filename := filepath.Join(dir, p.Name+extension)

	debug("configFilename:", filename)

	return filename
}

func (p *Pgosrv) saveConfig() error {
	debug("saveConfig")

	path := p.configFilename(configDir(), ".json")

	if !fileExists(path) {
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if isError(err) {
			return err
		}
	}

	ba, err := json.MarshalIndent(p, "", "  ")
	if isError(err) {
		return err
	}

	err = ioutil.WriteFile(path, ba, os.ModePerm)
	if isError(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) loadConfig(mustExist bool) error {
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
	if isError(err) {
		return err
	}

	err = json.Unmarshal(ba, p)
	if isError(err) {
		return err
	}

	return nil
}

func (p *Pgosrv) deleteConfig() error {
	debug("deleteConfig")

	path := p.configFilename(configDir(), ".json")

	if !fileExists(path) {
		return fmt.Errorf("configuration is not available or readable: %s", path)
	}

	err := os.RemoveAll(filepath.Dir(path))
	if isError(err) {
		return err
	}

	return nil
}

func run() error {
	banner()

	err := openLog()
	if isError(err) {
		return err
	}

	defer closeLog()

	if len(os.Args) < 2 || hasFlag("//?") {
		usage()

		return nil
	}

	debug("cmdline:", surroundWidth(os.Args, "\"", " "))

	p := &Pgosrv{}

	err = p.scanArgs()
	if isError(err) {
		return err
	}

	if p.Name == "" {
		return fmt.Errorf("missing service name")
	}

	debug("Service:", p.Name)

	//if p.LogPath != "" {
	//	if fileExists(p.LogPath) {
	//		var err error
	//
	//		logf, err = os.OpenFile(p.configFilename(p.LogPath, ".log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	//		if isError(err) {
	//			return err
	//		}
	//
	//		for _, l := range logs {
	//			logf.WriteString(l)
	//		}
	//
	//		defer logf.Close()
	//
	//		log.SetPrefix(p.LogPrefix + " ")
	//		log.SetOutput(io.MultiWriter(os.Stdout, logf))
	//	}
	//}

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
	if isError(run()) {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
