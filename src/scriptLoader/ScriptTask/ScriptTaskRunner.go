package ScriptTask

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	
	tpc "twoPC/Script2PC"
	sUtils "scriptLoader/utils"
	tUtils "twoPC/utils"
	
	"github.com/hashicorp/consul/api"
)

type scriptInfo struct {
	taskName     string
	action       string
	actionArgs   []string
	rollback     string
	rollbackArgs []string
	timeout      string
}

const Target = "finished with exit code 0"

func RunScriptTask(taskName string) {
	cln, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		fmt.Println("Create client failed. " + err.Error())
		os.Exit(0)
	}
	
	taskInfo := tpc.TaskInfo{TaskName: taskName}
	err = taskInfo.GetTask(cln)
	if err != nil {
		fmt.Println("Get task failed. " + err.Error())
		err1 := os.RemoveAll(taskInfo.TaskDir)
		if err1 != nil {
			fmt.Println(err1)
		}
		os.Exit(0)
	}
	
	nodeName, _ := cln.Agent().NodeName()
	err = tUtils.SetStatus(cln, tpc.TaskKVPrefix+taskInfo.TaskName+tpc.NodeStatus+nodeName, tpc.Running, 0)
	
	runInfo := make([]string, 0)
	cmd := exec.Command(taskInfo.Action, append(runInfo, taskInfo.ActionArgs...)...)
	out, err := cmd.Output()
	
	errB := err == nil
	errO := strings.Contains(string(out), Target)
	if errB && errO {
		tUtils.SetStatus(cln, tpc.TaskKVPrefix+taskInfo.TaskName+tpc.NodeStatus+nodeName, tpc.Success, 0)
	} else {
		if !errB {
			fmt.Println(err)
			tUtils.SetStatus(cln, tpc.TaskKVPrefix+taskInfo.TaskName+tpc.NodeStatus+nodeName, tpc.Fail, 0)
		}
		if !errO {
			fmt.Println(string(out))
			tUtils.SetStatus(cln, tpc.TaskKVPrefix+taskInfo.TaskName+tpc.NodeStatus+nodeName, tpc.Fail, 0)
		}
	}
	
	timeout, err := strconv.ParseUint(taskInfo.Timeout, 10, 64)
	if err != nil {
		timeout = tpc.DefaultTimeout
	}
	err = sUtils.CheckTaskStatus(cln, taskInfo.TaskName, timeout)
	if err != nil {
		fmt.Println("Task is aborted, start rollback.")
		err1 := tUtils.SetStatus(cln, tpc.TaskKVPrefix+taskInfo.TaskName+tpc.NodeStatus+nodeName, tpc.Fail, 0)
		if err1 != nil {
			fmt.Println(err1)
		}
		rollback(taskInfo.Rollback, taskInfo.RollbackArgs)
		err1 = tUtils.SetStatus(cln, tpc.TaskKVPrefix+taskInfo.TaskName+tpc.CheckedStatus+nodeName, tpc.Checked, 0)
		if err1 != nil {
			fmt.Println(err1)
		}
		err1 = os.RemoveAll(taskInfo.TaskDir)
		if err1 != nil {
			fmt.Println(err1)
		}
		os.Exit(0)
	}
	
	err = tUtils.SetStatus(cln, tpc.TaskKVPrefix+taskInfo.TaskName+tpc.CheckedStatus+nodeName, tpc.Checked, 0)
	if err != nil {
		fmt.Println(err)
	}
	err = os.RemoveAll(taskInfo.TaskDir)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(0)
}

func rollback(rollbackScript string, rollbackArgs []string) {
	rollInfo := make([]string, 0)
	cmd := exec.Command(rollbackScript, append(rollInfo, rollbackArgs...)...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Start rollback action faild! " + err.Error())
	}
	if !strings.Contains(string(out), Target) {
		fmt.Println("rollback action failed!\n" + string(out))
	}
}

