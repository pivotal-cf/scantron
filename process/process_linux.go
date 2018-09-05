// +build !windows

package process

import (
	"fmt"
	"github.com/keybase/go-ps"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pivotal-cf/scantron"
	"github.com/pivotal-cf/scantron/netstat"
)

func GetProcesses() ([]scantron.Process, error) {
	rawProcesses, err := ps.Processes()

	if err != nil {
		return nil, err
	}
	processes := []scantron.Process{}
	for _, rawProcess := range rawProcesses {
		pid := rawProcess.Pid()
		process := scantron.Process{
			CommandName: rawProcess.Executable(),
			PID:         pid,
			User: 			 GetUser(pid),
			Cmdline: GetCmdline(pid),
			Env: GetEnv(pid),
		}
		processes = append(processes, process)
	}

	return processes, nil
}

func GetPorts() ProcessPorts {
	bs, err := exec.Command("netstat", "-at", "-4", "-6", "--numeric-ports", "-u", "-p").Output()
	if err != nil {
		return nil
	}

	netstatPorts := netstat.ParseNetstatOutputForPort(string(bs))
	processPorts := []ProcessPort{}
	for _, np := range netstatPorts {
		processPorts = append(processPorts, ProcessPort{
			PID: np.PID,
			Port: np.Port,
		})
	}

	return processPorts
}

func GetUser(pid int) string {
	bs, err := exec.Command("ps", "-e", "-o", "uname:20=", "-f", strconv.Itoa(pid)).CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting user:", err)
		return ""
	}

	return strings.TrimSpace(string(bs))
}

func GetCmdline(pid int) []string {
	cmdline, err := readFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting cmdline:", err)
		return []string{}
	}

	return cmdline
}

func GetEnv(pid int) []string {
	env, err := readFile(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting env:", err)
		return []string{}
	}

	return env
}
