package cmd

import (
	"fmt"
	tunnelclient "github.com/TeaWeb/tunnel-client"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/files"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

type Shell struct {
	ShouldStop bool
}

func (this *Shell) Start() {
	// reset root directory
	this.resetRoot()

	// execute arguments
	if this.execArgs() {
		this.ShouldStop = true
		return
	}

	// write current pid
	files.NewFile(Tea.Root + Tea.DS + "bin" + Tea.DS + "pid").
		WriteString(fmt.Sprintf("%d", os.Getpid()))

	// log
	if len(os.Args) > 1 && !Tea.IsTesting() {
		fp, err := os.OpenFile(Tea.Root+"/logs/run.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(fp)
		} else {
			logs.Println("[error]" + err.Error())
		}
	}

	// start tunnel manager
	err := tunnelclient.SharedManager.Start()
	if err != nil {
		logs.Println("[error]" + err.Error())
	}

	// wait
	done := make(chan bool)
	<-done
}

// reset root
func (this *Shell) resetRoot() {
	if !Tea.IsTesting() {
		exePath, err := os.Executable()
		if err != nil {
			exePath = os.Args[0]
		}
		link, err := filepath.EvalSymlinks(exePath)
		if err == nil {
			exePath = link
		}
		fullPath, err := filepath.Abs(exePath)
		if err == nil {
			Tea.UpdateRoot(filepath.Dir(filepath.Dir(fullPath)))
		}
	}
	Tea.SetPublicDir(Tea.Root + Tea.DS + "web" + Tea.DS + "public")
	Tea.SetViewsDir(Tea.Root + Tea.DS + "web" + Tea.DS + "views")
	Tea.SetTmpDir(Tea.Root + Tea.DS + "web" + Tea.DS + "tmp")
}

// check command line arguments
func (this *Shell) execArgs() bool {
	if len(os.Args) == 1 {
		// check process pid
		proc := this.checkPid()
		if proc != nil {
			fmt.Println("TeaWeb Tunnel Client is already running, pid:", proc.Pid)
			return true
		}
		return false
	}
	args := os.Args[1:]
	if lists.ContainsAny(args, "?", "help", "-help", "h", "-h") {
		return this.execHelp()
	} else if lists.ContainsAny(args, "-v", "version", "-version") {
		return this.execVersion()
	} else if lists.ContainsString(args, "start") {
		return this.execStart()
	} else if lists.ContainsString(args, "stop") {
		return this.execStop()
	} else if lists.ContainsString(args, "restart") {
		return this.execRestart()
	} else if lists.ContainsString(args, "status") {
		return this.execStatus()
	}

	if len(args) > 0 {
		fmt.Println("Unknown command option '" + strings.Join(args, " ") + "', run './bin/teaweb-tunnel -h' to lookup the usage.")
		return true
	}
	return false
}

// command line helps
func (this *Shell) execHelp() bool {
	fmt.Println("TeaWeb Tunnel v" + tunnelclient.Version)
	fmt.Println("Usage:", "\n   ./bin/teaweb-tunnel [option]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -h", "\n     print this help")
	fmt.Println("  -v", "\n     print version")
	fmt.Println("  start", "\n     start the tunnel client in background")
	fmt.Println("  stop", "\n     stop the tunnel client")
	fmt.Println("  restart", "\n     restart the tunnel client")
	fmt.Println("  status", "\n     print tunnel client status")
	fmt.Println("")
	fmt.Println("To run the tunnel client in foreground:", "\n   ./bin/teaweb-tunnel")

	return true
}

// version
func (this *Shell) execVersion() bool {
	fmt.Println("TeaWeb Tunnel Client v"+tunnelclient.Version, "(build: "+runtime.Version(), runtime.GOOS, runtime.GOARCH+")")
	return true
}

// start the server
func (this *Shell) execStart() bool {
	proc := this.checkPid()
	if proc != nil {
		fmt.Println("TeaWeb Tunnel Client already started, pid:", proc.Pid)
		return true
	}

	cmd := exec.Command(os.Args[0], "background")
	err := cmd.Start()
	if err != nil {
		fmt.Println("TeaWeb Tunnel Client start failed:", err.Error())
		return true
	}
	fmt.Println("TeaWeb Tunnel Client started ok, pid:", cmd.Process.Pid)

	return true
}

// stop the server
func (this *Shell) execStop() bool {
	proc := this.checkPid()
	if proc == nil {
		fmt.Println("TeaWeb Tunnel Client not started")
		return true
	}

	err := proc.Kill()
	if err != nil {
		fmt.Println("TeaWeb Tunnel Client stop error:", err.Error())
		return true
	}

	files.NewFile(Tea.Root + "/bin/pid").Delete()
	fmt.Println("TeaWeb Tunnel Client stopped ok, pid:", proc.Pid)

	return true
}

// restart the server
func (this *Shell) execRestart() bool {
	proc := this.checkPid()
	if proc != nil {
		err := proc.Kill()
		if err != nil {
			fmt.Println("TeaWeb Tunnel Client stop error:", err.Error())
			return true
		}
	}

	cmd := exec.Command(os.Args[0])
	err := cmd.Start()
	if err != nil {
		fmt.Println("TeaWeb Tunnel Client restart failed:", err.Error())
		return true
	}
	fmt.Println("TeaWeb Tunnel Client restarted ok, pid:", cmd.Process.Pid)

	return true
}

// server status
func (this *Shell) execStatus() bool {
	proc := this.checkPid()
	if proc == nil {
		fmt.Println("TeaWeb Tunnel Client not started yet")
	} else {
		fmt.Println("TeaWeb Tunnel Client is running, pid:" + fmt.Sprintf("%d", proc.Pid))
	}
	return true
}

// check process pid
func (this *Shell) checkPid() *os.Process {
	// check pid file
	pidFile := files.NewFile(Tea.Root + "/bin/pid")
	if !pidFile.Exists() {
		return nil
	}
	pidString, err := pidFile.ReadAllString()
	if err != nil {
		return nil
	}
	pid := types.Int(pidString)

	if pid <= 0 {
		return nil
	}

	// if pid equals current pid
	if pid == os.Getpid() {
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil || proc == nil {
		return nil
	}

	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		return nil
	}

	// ps?
	ps, err := exec.LookPath("ps")
	if err != nil {
		return proc
	}

	cmd := exec.Command(ps, "-p", pidString, "-o", "command=")
	output, err := cmd.Output()
	if err != nil {
		return proc
	}

	if len(output) == 0 {
		return nil
	}

	outputString := string(output)
	index := strings.LastIndex(outputString, "/")
	if index > -1 {
		outputString = outputString[index+1:]
	}
	index2 := strings.LastIndex(outputString, "\\")
	if index2 > 0 {
		outputString = outputString[index2+1:]
	}
	if strings.Contains(outputString, "teaweb-tunnel") {
		return proc
	}

	return nil
}
