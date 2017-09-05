package main

import (
	"fmt"
	"twoPC/Script2PC"
	"flag"
	"strings"
)

type ScriptArgs []string

func (sa *ScriptArgs) String() string {
	return fmt.Sprint(*sa)
}

func (sa *ScriptArgs) Set(value string) error {
	for _, arg := range strings.Split(value, ",") {
		*sa = append(*sa, arg)
	}
	return nil
}

func main() {
	var nodes ScriptArgs
	flag.Var(&nodes, "n", "Comma-separated list of strings of node list.")
	var scriptArgs ScriptArgs
	flag.Var(&scriptArgs, "aa", "Comma-separated list of strings of action script args.")
	var rollbackArgs ScriptArgs
	flag.Var(&rollbackArgs, "ra", "Comma-separated list of strings of rollback script args.")
	taskName := flag.String("tn", Script2PC.DefaultTaskName, "Task Name.")
	action := flag.String("a", "/root/demo/action.sh", "Action script path.")
	rollback := flag.String("r", "/root/demo/rollback.sh", "Rollback script path.")
	timeout := flag.String("to", "20", "Task timeout")
	flag.Parse()
	var task Script2PC.TaskInfo
	task.New(*taskName, nodes, *action, scriptArgs, *rollback, rollbackArgs, *timeout)
	err := task.RunAction()
	if err != nil {
		fmt.Println(err)
	}
}
