package main

import (
	"flag"
	
	tpc "twoPC/Script2PC"
	st "scriptLoader/ScriptTask"
)

type scriptInfo struct {
	taskName     string
	action       string
	actionArgs   []string
	rollback     string
	rollbackArgs []string
	timeout      string
}

func main() {
	taskName := flag.String("t", tpc.DefaultTaskName, "The task name.")
	flag.Parse()
	st.RunScriptTask(*taskName)
}
