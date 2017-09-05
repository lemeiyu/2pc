package two_pc_test

import (
	"testing"
	tpc "twoPC/Script2PC"
	"twoPC/utils"
	"github.com/hashicorp/consul/api"
	"fmt"
	"os"
)

func Test_PutGet(t *testing.T) {
	nodes := []string{"Qi-MacBook-Pro"}
	scriptArgs := []string{"args1", "args2", "args3", "args4", "-done"}
	rollbackArgs := make([]string, 0)
	task := tpc.TaskInfo{
		TaskName:       "Demo",
		Nodes:          nodes,
		Action:         "/Users/leqi/Dev/DevTemp/demo/task.sh",
		TaskDir:        "/Users/leqi/Dev/DevTemp/demo/rcv",
		ActionSuffix:   ".sh",
		ActionArgs:     scriptArgs,
		Rollback:       "/Users/leqi/Dev/DevTemp/demo/rollback.sh",
		//RollbackDir:    "/Users/leqi/Dev/DevTemp/demo/rcv",
		RollbackSuffix: ".sh",
		RollbackArgs:   rollbackArgs,
		Timeout:        "2000"}
	cln, _ := api.NewClient(api.DefaultConfig())
	cln.KV().DeleteTree(tpc.TaskKVPrefix, nil)
	task.PutTask(cln)
	dtask := tpc.TaskInfo{TaskName: "Demo"}
	dtask.GetTask(cln)
	fmt.Print(task.TaskName, task.Action, task.TaskDir, task.Rollback, task.Timeout)
	os.RemoveAll(dtask.TaskDir)
}

func Test_RunScriptTest(t *testing.T) {
	cln, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return
	}
	nodes := []string{"Qi-MacBook-Pro"}
	scriptArgs := []string{"-none"}
	rollbackArgs := []string{"-none"}
	taskName := "Demo"
	action := "/Users/leqi/Dev/DevTemp/demo/task.sh"
	rollback := "/Users/leqi/Dev/DevTemp/demo/rollback.sh"
	timeout := "2000"
	var task tpc.TaskInfo
	task.New(taskName, nodes, action, scriptArgs, rollback, rollbackArgs, timeout)
	utils.DeleteInfo(cln, tpc.TaskKVPrefix+taskName)
	err = task.RunAction()
	if err != nil {
		fmt.Println(err)
	}
}

func Test_Suffix(t *testing.T) {
	s := "/Users/124/ddd/k.d/ss.sh"
	s = utils.FileSuffix(s)
	fmt.Print(s)
}

func Test_RemoveAll(t *testing.T) {
	err := os.RemoveAll("/Users/leqi/Dev/DevTemp/demo/rcv")
	if err != nil {
		fmt.Println(err)
	}
}