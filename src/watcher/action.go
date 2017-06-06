package watcher

import (
	"github.com/coreos/etcd/client"

	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"utils"
	"utils/xlog"
)

var (
	ExecCmdTimeout = 5 * time.Second
)

func handleProAction(proPrefix string, watcher *Watcher, exitCh chan bool) {
	xlog.Debug("watch prefix :%v", fmt.Sprintf("%v/%v", proPrefix, EtcdWatchNode))
	for {
		select {
		case resp, ok := <-watcher.respProCh:
			if !ok {
				xlog.Warn("recv from project resp chan failed, channel may be closed. node:%v", fmt.Sprintf("%v/%v", proPrefix, EtcdWatchNode))
				goto exit
			}

			switch resp.Action {
			case "get":
				// TODO: noting
			case "create", "set", "update":
				go setAction(watcher, resp)
			case "delete":
				go deleteAction(watcher, resp)
			}
		case <-exitCh:
			goto exit
		}
	}

exit:
	xlog.Debug("cannel watch prefix :%v", fmt.Sprintf("%v/%v", proPrefix, EtcdWatchNode))
}

func setAction(w *Watcher, resp *client.Response) {
	var (
		err       error
		filename  string
		config    Config
		beforeCmd Cmd
		afterCmd  Cmd
	)
	// recover for the last time
	defer func() {
		if revErr := recover(); err != nil {
			xlog.Fatal("setAction: recover is err, err:%v", revErr)
		}
	}()

	// callback
	defer func() {
		if len(config.Callback) == 0 {
			return
		}

		var respErr error
		var code int
		var msg string
		if err != nil {
			code = http.StatusInternalServerError
			msg = err.Error()
		} else {
			code = http.StatusOK
		}

		if beforeCmd.Err != nil {
			beforeCmd.Msg = beforeCmd.Err.Error()
		}
		if afterCmd.Err != nil {
			afterCmd.Msg = afterCmd.Err.Error()
		}
		response := &Response{
			Code:      code,
			Msg:       msg,
			MD5:       utils.GetMD5Hash(resp.Node.Value),
			BeforeCmd: beforeCmd,
			AfterCmd:  afterCmd,
		}
		response.Action = resp.Action
		respErr = response.Callback(config.Callback)
		if respErr != nil {
			xlog.Fatal("setAction callback: response.Callback is err, err:%v", respErr)
		}
	}()

	// get project config from etcd
	proPrefix := trimProPrefix(resp.Node.Key, w.cfg.Prefix)
	prefix, conf, err := w.getConfig(proPrefix)
	if err != nil {
		xlog.Warn("setAction getConfig: node:%v, err:%v", prefix, err)
		return
	}
	err = json.Unmarshal(conf, &config)
	if err != nil {
		xlog.Warn("setAction json.Unmarshal: node:%v, err:%v", prefix, err)
		return
	}
	err = config.checkConfig()
	if err != nil {
		xlog.Warn("setAction checkConfig: node:%v, err:%v", prefix, err)
		return
	}

	strarr := strings.Split(resp.Node.Key, "/")
	filename = strarr[len(strarr)-1]
	if len(filename) == 0 {
		xlog.Warn("setAction: file name is null, action:%v", resp.Action)
		return
	}

	if !utils.FileExists(config.DeployPath) {
		err = os.MkdirAll(config.DeployPath, 0755)
		if err != nil {
			xlog.Warn("setAction: os.Mkdir is err, action:%v, err:%v", resp.Action, err)
			return
		}
	}

	path := config.DeployPath
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	file := path + filename
	xlog.Debug("setAction: write to %v", file)

	// publish before
	beforeCmd.Success, beforeCmd.Out, beforeCmd.Err = runCmd(config.BeforeCmd)

	err = utils.FileWrite(file, &resp.Node.Value)
	if err != nil {
		xlog.Warn("setAction: utils.FileWrite is err, action:%v, err:%v", resp.Action, err)
	}

	// publish after
	afterCmd.Success, afterCmd.Out, afterCmd.Err = runCmd(config.AfterCmd)

	return
}

func deleteAction(w *Watcher, resp *client.Response) {
	var (
		err       error
		filename  string
		config    Config
		beforeCmd Cmd
		afterCmd  Cmd
	)

	// recover for the last time
	defer func() {
		if revErr := recover(); err != nil {
			xlog.Fatal("deleteAction: recover is err, err:%v", revErr)
		}
	}()

	// callback
	defer func() {
		if len(config.Callback) == 0 {
			return
		}

		var respErr error
		var code int
		var msg string
		if err != nil {
			code = http.StatusInternalServerError
			msg = err.Error()
		} else {
			code = http.StatusOK
		}

		if beforeCmd.Err != nil {
			beforeCmd.Msg = beforeCmd.Err.Error()
		}
		if afterCmd.Err != nil {
			afterCmd.Msg = afterCmd.Err.Error()
		}
		response := &Response{
			Code:      code,
			Msg:       msg,
			MD5:       "",
			BeforeCmd: beforeCmd,
			AfterCmd:  afterCmd,
		}
		response.Action = resp.Action
		respErr = response.Callback(config.Callback)
		if respErr != nil {
			xlog.Fatal("deleteAction callback: response.Callback is err, err:%v", respErr)
		}
	}()

	// get project config from etcd
	proPrefix := trimProPrefix(resp.Node.Key, w.cfg.Prefix)
	prefix, conf, err := w.getConfig(proPrefix)
	if err != nil {
		xlog.Warn("deleteAction getConfig: node:%v, err:%v", prefix, err)
		return
	}
	err = json.Unmarshal(conf, &config)
	if err != nil {
		xlog.Warn("deleteAction json.Unmarshal: node:%v, err:%v", prefix, err)
		return
	}

	strarr := strings.Split(resp.Node.Key, "/")
	filename = strarr[len(strarr)-1]
	if len(filename) == 0 {
		xlog.Warn("deleteAction: filename is null, action:%v", resp.Action)
		return
	}

	path := config.DeployPath
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	file := path + filename
	if !utils.FileExists(file) {
		xlog.Warn("deleteAction: file doesn't exist, action:%v", resp.Action)
		return
	}

	// publish before
	beforeCmd.Success, beforeCmd.Out, beforeCmd.Err = runCmd(config.BeforeCmd)

	// when backup dir is null that will remove the config file
	// when backup dir is seted that will backup the config file to backup dir
	if len(config.BackupDir) == 0 {
		err = os.Remove(file)
		if err != nil {
			xlog.Warn("deleteAction: os.Remove is err, action:%v, err:%v", resp.Action, err)
			return
		}
		xlog.Debug("deleteAction: remove file %v", file)
	} else {
		isExists := utils.FileExists(config.BackupDir)
		if !isExists {
			err = os.MkdirAll(config.BackupDir, 0755)
			if err != nil {
				xlog.Warn("deleteAction: os.MkdirAll is err, action:%v, err:%v", resp.Action, err)
				return
			}
		}
		isdir, err := utils.IsDir(config.BackupDir)
		if err != nil {
			xlog.Warn("deleteAction: utils.IsDir is err, action:%v, err:%v", resp.Action, err)
			return
		}
		if !isdir {
			xlog.Warn("deleteAction: config.BackupDir is not a dir, action:%v, err:%v", resp.Action, err)
		}
		timestamp := time.Now().Unix()
		tm := time.Unix(timestamp, 0)
		newfile := fmt.Sprintf("%v/%v_watcherbackup_%v_%v", config.BackupDir, filename, tm.Format(TimeFormat), rand.Int63())
		err = os.Rename(file, newfile)
		if err != nil {
			xlog.Warn("deleteAction: os.Rename is err, action:%v, err:%v", resp.Action, err)
		}
		xlog.Debug("deleteAction: backup file %v to %v", file, newfile)
	}

	// publish after
	afterCmd.Success, afterCmd.Out, afterCmd.Err = runCmd(config.AfterCmd)

	return
}

func runCmd(cmd string) (bool, string, error) {
	if len(cmd) == 0 {
		return true, "", nil
	}

	cmdArgs := strings.Split(cmd, " ")
	name := cmdArgs[0]
	args := cmdArgs[1:]

	cmdSuccess, out, cmdErr := utils.Command(ExecCmdTimeout, name, args...)
	cmdOut := string(out)
	return cmdSuccess, cmdOut, cmdErr
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
