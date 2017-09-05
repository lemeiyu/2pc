package Script2PC

import (
	"fmt"
	"time"
	"strconv"
	
	"twoPC/utils"
	
	"github.com/hashicorp/consul/api"
)

const (
	ScriptLoaderPath     = "scriptLoader"
	TaskKVPrefix         = "Script2PC/ScriptTasks/"
	DefaultTaskName      = "Default"
	DefaultTimeout       = 20
	TaskStatus           = "/Status/TaskStatus"
	NodeStatus           = "/Status/NodeStatus/"
	CheckedStatus        = "/Status/CheckedStatus/"
	TaskInfoKey          = "/TaskInfo"
	ActionScript         = "/TaskInfo/FileInfo/ActionScript"
	ActionScriptSuffix   = "/TaskInfo/FileInfo/ActionScriptSuffix"
	RollbackScript       = "/TaskInfo/FileInfo/RollbackScript"
	RollbackScriptSuffix = "/TaskInfo/FileInfo/RollbackScriptSuffix"
	ActionArgs           = "/TaskInfo/Args/ActionArgs/"
	RollbackArgs         = "/TaskInfo/Args/RollbackArgs/"
	Timeout              = "/TaskInfo/Timeout"
	
	Prepare = "prepare"
	Commit  = "commit"
	Abort   = "abort"
	Running = "running"
	Success = "success"
	Fail    = "fail"
	Checked = "yes"
)

type TaskInfo struct {
	TaskName       string
	Nodes          []string
	Action         string
	ActionArgs     []string
	ActionSuffix   string //extract file suffix and attach to "action"
	Rollback       string
	RollbackArgs   []string
	RollbackSuffix string //extract file suffix and attack to "rollback"
	TaskDir        string //will create a temp dir
	Timeout        string
}

func (t *TaskInfo) New(taskName string, nodes []string, action string, actArg []string, rollback string, rollArgs []string, timeout string) {
	if taskName == "" {
		t.TaskName = DefaultTaskName
	} else {
		t.TaskName = taskName
	}
	t.Nodes = nodes
	t.Action = action
	if actArg == nil {
		t.ActionArgs = make([]string, 0)
	} else {
		t.ActionArgs = actArg
	}
	t.ActionSuffix = utils.FileSuffix(action)
	t.Rollback = rollback
	if rollArgs == nil {
		t.RollbackArgs = make([]string, 0)
	} else {
		t.RollbackArgs = rollArgs
	}
	t.RollbackSuffix = utils.FileSuffix(rollback)
	t.Timeout = timeout
}

//If any error thrown, run DeleteInfo manually
//or no task will be denied
func (t *TaskInfo) RunAction() error {
	cln, err := api.NewClient(api.DefaultConfig())
	
	if formerTask, _, _ := cln.KV().Keys(TaskKVPrefix, "/", nil); formerTask != nil {
		return fmt.Errorf("Former task is running.")
	}
	
	err = t.PutTask(cln)
	if err != nil {
		return err
	}
	
	timeFlag := uint64(time.Now().Unix())
	err = utils.SetStatus(cln, TaskKVPrefix+t.TaskName+TaskStatus, Prepare, timeFlag)
	if err != nil {
		return err
	}
	
	err = utils.ConsulExec(t.Nodes, ScriptLoaderPath, "-t", t.TaskName)
	if err != nil {
		err1 := utils.SetStatus(cln, TaskKVPrefix+t.TaskName+TaskStatus, Abort, timeFlag)
		if err1 != nil {
			return err1
		}
		return fmt.Errorf("Start task failed: %s", err.Error())
	}
	
	timeout, err := strconv.Atoi(t.Timeout)
	if err != nil {
		timeout = DefaultTimeout
	}
	err = CheckMultiStatus(cln, t.Nodes, t.TaskName, timeout, NodeStatus)
	if err != nil {
		err1 := utils.SetStatus(cln, TaskKVPrefix+t.TaskName+TaskStatus, Abort, timeFlag)
		if err1 != nil {
			return err1
		}
		return fmt.Errorf("Task failed: %s", err.Error())
	}
	
	err = utils.SetStatus(cln, TaskKVPrefix+t.TaskName+TaskStatus, Commit, timeFlag)
	if err != nil {
		return err
	}
	
	err = CheckMultiStatus(cln, t.Nodes, t.TaskName, DefaultTimeout, CheckedStatus)
	if err != nil {
		err1 := utils.DeleteInfo(cln, TaskKVPrefix+t.TaskName)
		if err1 != nil {
			return err1
		}
		return err
	}
	
	err = utils.DeleteInfo(cln, TaskKVPrefix+t.TaskName)
	if err != nil {
		return err
	}
	
	return nil
}

//CheckMultiStatus is used to check NodeStatus and CheckedStatus. And as the key is a parameter of the function, it also
//can be used to check other kv pair with a specified key.
func CheckMultiStatus(cln *api.Client, nodes []string, taskName string, timeout int, key string) error {
	
	t := time.Now()
	stop := true
	
	for {
		time.Sleep(500 * time.Millisecond)
		status, _, err := cln.KV().List(TaskKVPrefix+taskName+key, nil)
		if d := time.Since(t); d < (time.Duration(timeout) * time.Second) {
			//status, _, err := cln.KV().List(TaskKVPrefix+taskName+key, nil)
			if err != nil {
				continue
			}
			if len(status) != len(nodes) {
				continue
			} else {
				stop, err = statusChecker(status)
				if err != nil {
					return err
				}
			}
			if stop {
				break
			}
			stop = true
		} else {
			if key == NodeStatus {
				return fmt.Errorf("Task timeout, try to rollback.")
			}
			
			if err != nil || status == nil {
				return fmt.Errorf("Fetch commit status failed, check connection.")
			}
			errMsg := ""
			for _, s := range status {
				errMsg = errMsg + s.Key + "; "
			}
			return fmt.Errorf("These nodes has committed/aborted: %s, but other(s) not.", errMsg)
			
		}
		
	}
	return nil
}

func statusChecker(status api.KVPairs) (bool, error) {
	for idx, s := range status {
		if m := string(s.Value); m == Success || m == Checked {
			if idx == len(status)-1 {
				break
			}
			continue
		} else if m == Running {
			return false, nil
		} else if m == Fail {
			return true, fmt.Errorf("Task failed on node: %s, try to rollback.", status[idx].Key)
		}
	}
	return true, nil
}
