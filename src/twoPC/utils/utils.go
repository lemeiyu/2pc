package utils

import (
	"os"
	"fmt"
	"sync"
	"os/exec"
	"strings"
	
	"github.com/hashicorp/consul/api"
)

func ReadFromFile(filePath string) ([]byte, error) {
	scriptFile, err := os.Open(filePath)
	defer scriptFile.Close()
	if err != nil {
		return nil, fmt.Errorf("Open Action file: %s failed, %s", filePath, err.Error())
	}
	
	var outputBuf []byte
	buf := make([]byte, 1024)
	for {
		n, _ := scriptFile.Read(buf)
		
		if n == 0 {
			break
		} else {
			outputBuf = append(outputBuf, buf[: n]...)
		}
		
	}
	return outputBuf, nil
}

func WriteToFile(filePath string, byteData []byte) error {
	scriptFile, err := os.Create(filePath)
	defer scriptFile.Close()
	if err != nil {
		return fmt.Errorf("Create Action file: %s failed, %s", filePath, err.Error())
	}
	
	//_, err = scriptFile.WriteString(string(byteData))
	_, err = scriptFile.Write(byteData)
	if err != nil {
		return fmt.Errorf("Write Action file: %s failed, %s", filePath, err.Error())
	}
	scriptFile.Chmod(0724)
	
	return nil
}

//A wrapper for 'consul exec', but will not wait for the command return.
func ConsulExec(nodes []string, progname string, arg ...string) error {
	var wg sync.WaitGroup
	wg.Add(len(nodes))
	
	res := make([]error, len(nodes))
	for idx, node := range nodes {
		go func(i int, n string) {
			defer wg.Done()
			args := []string{"exec", "-node", n, progname}
			cmd := exec.Command("consul", append(args, arg...)...)
			err := cmd.Start()
			res[i] = err
		}(idx, node)
	}
	
	wg.Wait()
	
	for idx, r := range res {
		if r != nil {
			return fmt.Errorf("%s: %s", nodes[idx], r.Error())
		}
	}
	
	return nil
}

//SetStatus is used to set a (status) key to a specified value and flag.
func SetStatus(cln *api.Client, key string, status string, flag uint64) error {
	p := api.KVPair{
		Key:   key,
		Value: []byte(status),
		Flags: flag}
	
	_, err := cln.KV().Put(&p, nil)
	if err != nil {
		return fmt.Errorf("Set %s to \"%s\" failed! %s", key, status, err.Error())
	}
	return nil
}


func FileSuffix(filePath string) string {
	if d := strings.LastIndex(filePath, "."); d != -1 {
		return filePath[d: ]
	} else {
		return ""
	}
}

func DeleteInfo(cln *api.Client, key string) error {
	_, err := cln.KV().DeleteTree(key, nil)
	if err != nil {
		return fmt.Errorf("Delete task info failed. %s", err.Error())
	}
	return nil
}
