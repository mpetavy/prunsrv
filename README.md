## PRUNSRV

### Project description

Just developed to get around some Apache PRUNSRV problems with the latest version 1.3.1

Your welcome to use this solution.

Most of the original parameters are working the same.
Please consider the original PRUNSRV which can be found at
https://commons.apache.org/proper/commons-daemon/procrun.html

### Differences to Apache PRUNSRV

* Works only with "Java" mode (no "jvm" or "exe" mode supported)
* Calls the static "StartClass" "main" static method with "StartMethod" argument
* Calls the static "StopClass" "main" static method with "StopMethod" argument
* Executes Java executable as separated processes, so no "jvm.dll" integration
* Calls Java static methods for "Start" and "Stop"
* No dependencies on naming of Java static methods
* Stores service configuration as JSON file to "ProgramData/prunsrv" (Windows) or "etc" (*nix)

### Supported commands

| Command              | Description                                                      |
|----------------------| ---------------------------------------------------------------- |
| //TS//<service name> | Run (test) the service in your console                           |
| //RS//<service name> | Used by the OS service manager to start the service as a service |
| //ES//<service name> | Start the service                                                |
| //SS//<service name> | Stop the service                                                 |
| //IS//<service name> | Install the services in the OS service manager                   |
| //US//<service name> | Uninstall the services in the OS service manager                 |
| //PS//<service name> | Print the current saved configuration in callable format         |
| //?                  | Shows help                                                       |

### Support parameters

| Command         | Default | Description                                                         |
| --------------- | ------- | ------------------------------------------------------------------- |
| Description     |         | Description of service                                              |
| DisplayName     |         | Service name                                                        |
| StartPath       |         | Working directory of the Java executable which executes the service |
| Startup         | manual  | "auto", "manual", "disabled" service startup mode                   |
| JavaHome        |         | Path to the Java runtime to use                                     |
| JvmOptions      |         | Java system properties to set as Java "-D" parameters               |
| Classpath       |         | Classpath to use for the Java "-cp" parameter                       |
| JvmMx           |         | Java options "-Xmx"                                                 |
| JvmMs           |         | Java options "-Xms"                                                 |
| JvmSs           |         | Java options "-Xss"                                                 |
| StartClass      | Service | FQDN of the Java class which starts the service                     |
| StopClass       | Service | FQDN of the Java class which starts the service                     |
| StartMethod     | start   | Name of the static class method to call to start the service        |
| StopMethod      | stop    | Name of the static class method to call to stop the service         |
| StopTimeout     | 20      | Timeout in seconds after that the service is terminated             |
| LogPath         |         | Path to PRUNSRV log file                                            |
| LogLevel        | info    | "info" or "debug" level                                             |
| LogPrefix       |         | prefix to be used before each line on log                           |
| ServiceUser     |         | Username of the user under which service is run                     |
| ServicePassword |         | Password of the user under which service is run                     |
| PidFile         |         | Path to store the serice PID                                        |

### Samples

#### Install a service to OS service manager

    "prunsrv" ^
    "//TS//TestService" ^
    "--Description=Description of Test Service" ^
    "--DisplayName=TestService" ^
    "--StartPath=D:\java\myapp\myapp-server-parent\myapp-server" ^
    "--Startup=auto" ^
    "--JavaHome=c:\Program Files\AdoptOpenJDK\jdk-11.0.8.10-hotspot" ^
    "--JvmOptions=-Dfile.encoding=UTF-8" ^
    "++JvmOptions=-Dspring.config.location=d:\java\myapp\myapp-server-parent\myapp-server\target\classes\config\myapp.yml" ^
    "--Classpath=d:\java\myapp\myapp-server-parent\myapp-server\target\myapp-server-2.1.0-SNAPSHOT.jar" ^
    "--JvmMx=1024m" ^
    "--JvmMs=" ^
    "--JvmSs=" ^
    "--StartClass=de.zeiss.myapp.server.Service" ^
    "--StopClass=de.zeiss.myapp.server.Service" ^
    "--StartMethod=start" ^
    "--StopMethod=stop" ^
    "--StopTimeout=20" ^
    "--LogPath=d:\java\myapp\myapp-server-parent\myapp-server\log" ^
    "--LogLevel=Debug" ^
    "--LogPrefix=myapp" ^
    "--ServiceUser=.\johndoe" ^
    "--ServicePassword=supersecret" ^
    "--PidFile=d:\java\myapp\myapp-server-parent\myapp-server\log\myapp.pidfile"

#### Remove service to OS service manager

    prunsrv //DS//TestService

#### Test a service on current console

    prunsrv //TS//TestService


### License

All software is copyright and protected by the Apache License, Version 2.0.
https://www.apache.org/licenses/LICENSE-2.0