package utils

import (
	"time"
	"fmt"
	
	tpc "twoPC/Script2PC"
	
	"github.com/hashicorp/consul/api"
)

func CheckTaskStatus(cln *api.Client, taskName string, timeout uint64) error {
	key := tpc.TaskKVPrefix + taskName + tpc.TaskStatus
	for {
		time.Sleep(500 * time.Millisecond)
		status, _, err := cln.KV().Get(key, nil)
		if err != nil {
			continue
		}
		d := uint64(time.Now().Unix()) - status.Flags
		if d < timeout {
			if string(status.Value) == tpc.Commit {
				break
				//return nil
			} else if string(status.Value) == tpc.Abort {
				return fmt.Errorf("Task has been aborted.")
			} else {
				continue
			}
		} else {
			if string(status.Value) == tpc.Commit {
				break
			} else {
				return fmt.Errorf("Task timeout.")
			}
		}
	}
	return nil
}