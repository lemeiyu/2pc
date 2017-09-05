package Script2PC

import (
	"fmt"
	"strconv"
	"io/ioutil"
	"strings"
	
	"twoPC/utils"
	
	"github.com/hashicorp/consul/api"
)

const (
	ActionDownload   = "/action"
	RollbackDownload = "/rollback"
)

type FileSender struct {
	TaskName string
	FilePath string
	FileKey  string
}

func (f *FileSender) New(taskName string, filePath string, fileKey string) {
	f.TaskName = taskName
	f.FilePath = filePath
	f.FileKey = fileKey
}

//PutTask is used to put all the information about the task on kv storage.
func (t *TaskInfo) PutTask(cln *api.Client) error {
	taskPrefix := TaskKVPrefix + t.TaskName
	pairs := api.KVPairs{}
	as, err := utils.ReadFromFile(t.Action)
	if err != nil {
		return fmt.Errorf("Read action script failed, %s", err.Error())
	}
	pairs = append(pairs, &api.KVPair{Key: taskPrefix + ActionScript, Value: as})
	pairs = append(pairs, &api.KVPair{Key: taskPrefix + ActionScriptSuffix, Value: []byte(t.ActionSuffix)})
	//pairs = append(pairs, &api.KVPair{Key: taskPrefix + ActionScriptDir, Value: []byte(t.TaskDir)})
	
	rs, err := utils.ReadFromFile(t.Rollback)
	if err != nil {
		return fmt.Errorf("Read rollback script failed, %s", err.Error())
	}
	pairs = append(pairs, &api.KVPair{Key: taskPrefix + RollbackScript, Value: rs})
	pairs = append(pairs, &api.KVPair{Key: taskPrefix + RollbackScriptSuffix, Value: []byte(t.RollbackSuffix)})
	//pairs = append(pairs, &api.KVPair{Key: taskPrefix + RollbackScriptDir, Value: []byte(t.RollbackDir)})
	for idx, arg := range t.ActionArgs {
		pairs = append(pairs, &api.KVPair{Key: taskPrefix + ActionArgs + "args" + strconv.Itoa(idx), Value: []byte(arg)})
	}
	
	for idx, arg := range t.RollbackArgs {
		pairs = append(pairs, &api.KVPair{Key: taskPrefix + RollbackArgs + "args" + strconv.Itoa(idx), Value: []byte(arg)})
	}
	
	pairs = append(pairs, &api.KVPair{Key: taskPrefix + Timeout, Value: []byte(t.Timeout)})
	
	for _, pair := range pairs {
		_, err := cln.KV().Put(pair, nil)
		if err != nil {
			return fmt.Errorf("Put %s failed", string(pair.Value))
		}
	}
	return nil
}

//GetTask will Get all kv pairs under the task key prefix and check if all required info was downloaded. if so the info
//will be assigned to the proper variable
func (t *TaskInfo) GetTask(cln *api.Client) error {
	taskPrefix := TaskKVPrefix + t.TaskName
	
	tmpDir, err := ioutil.TempDir("", "Script2PC")
	if err != nil {
		return fmt.Errorf("Create task dir failed, %s", err.Error())
	}
	t.TaskDir = tmpDir
	
	pairs, _, err := cln.KV().List(taskPrefix+TaskInfoKey, nil)
	if err != nil {
		return fmt.Errorf("Get task failed, %s", err.Error())
	}
	
	var pAction, pActionSuffix, pRollback, pRollbackSuffix, pTimeout *api.KVPair
	var pActionArgs, pRollbackArgs api.KVPairs
	
	for _, pair := range pairs {
		if pair.Key == taskPrefix+ActionScriptSuffix {
			pActionSuffix = pair
		} else if pair.Key == taskPrefix+ActionScript {
			pAction = pair
		} else if pair.Key == taskPrefix+RollbackScriptSuffix {
			pRollbackSuffix = pair
		} else if pair.Key == taskPrefix+RollbackScript {
			pRollback = pair
		} else if pair.Key == taskPrefix+Timeout {
			pTimeout = pair
		} else if strings.Contains(pair.Key, taskPrefix+ActionArgs) {
			pActionArgs = append(pActionArgs, pair)
		} else if strings.Contains(pair.Key, taskPrefix+RollbackArgs) {
			pRollbackArgs = append(pRollbackArgs, pair)
		}
	}
	if pActionSuffix != nil {
		t.ActionSuffix = string(pActionSuffix.Value)
	} else {
		return fmt.Errorf("Get action script suffix failed.")
	}
	if pAction != nil {
		t.Action = tmpDir + ActionDownload + t.ActionSuffix
	} else {
		return fmt.Errorf("Get action script failed.")
	}
	if pRollbackSuffix != nil {
		t.RollbackSuffix = string(pRollbackSuffix.Value)
	} else {
		return fmt.Errorf("Get rollback script suffix failed.")
	}
	if pRollback != nil {
		t.Rollback = tmpDir + RollbackDownload + t.RollbackSuffix
	} else {
		return fmt.Errorf("Get rollback script failed.")
	}
	if pTimeout != nil {
		t.Timeout = string(pTimeout.Value)
	} else {
		return fmt.Errorf("Get timeout failed.")
	}
	
	t.ActionArgs = extractArgs(pActionArgs, taskPrefix+ActionArgs)
	t.RollbackArgs = extractArgs(pRollbackArgs, taskPrefix+RollbackArgs)
	
	err = utils.WriteToFile(tmpDir+ActionDownload+t.ActionSuffix, pAction.Value)
	if err != nil {
		return fmt.Errorf("Write action script failed, %s", err.Error())
	}
	err = utils.WriteToFile(tmpDir+RollbackDownload+t.ActionSuffix, pRollback.Value)
	if err != nil {
		return fmt.Errorf("Write rollback script failed, %s", err.Error())
	}
	
	return nil
}

func extractArgs(pairs api.KVPairs, prefix string) []string {
	l := len(prefix + "args")
	res := make([]string, len(pairs))
	for _, pair := range pairs {
		key := pair.Key[l:]
		idx, _ := strconv.Atoi(key)
		res[idx] = string(pair.Value)
	}
	return res
}


