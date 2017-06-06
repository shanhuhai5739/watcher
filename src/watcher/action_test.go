package watcher

import (
	"fmt"
	"github.com/coreos/etcd/client"
	"log"
	"testing"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"utils"
	"utils/xlog"
)

var (
	w *Watcher

	// etcd path
	prefix         string
	proPrefix      string
	proWatchPrefix string
	proConfPrefix  string
	hostname       = "corallocalhost"
	proName        = "rsyslog"
	config         Config
	proConfContent = `{"deployPath": "/tmp/watcher", "backupDir":"/tmp/backup", "beforeCmd":"echo before", "afterCmd":"echo after", "callback": "http://127.0.0.1:9090/callback"}`

	// config path
	ngxName = "nginx.conf"
	ngxConf = `
		proxy_connect_timeout    120;
		proxy_read_timeout       120;
		proxy_send_timeout       120;
		proxy_buffer_size        16k;
		proxy_buffers            4 64k;
		proxy_busy_buffers_size 128k;
		proxy_temp_file_write_size 128k;`

	cfg = Cfg{
		Endpoints:   "localhost:2379",
		DialTimeout: 5 * time.Second,
		Hostname:    hostname,
		Username:    "",
		Password:    "",
		Prefix:      "/watcher/unittest/",
		Force:       true,
	}
)

func TestCreateAction(t *testing.T) {
	var err error

	// exec create
	ngxConfPerfix := fmt.Sprintf("%s/%s", proWatchPrefix, ngxName)
	err = w.client.Create(ngxConfPerfix, []byte(ngxConf))
	if err != nil {
		t.Fatal(err)
	}

	// wait a moment
	time.Sleep(2 * time.Second)
	filename := fmt.Sprintf("%s/%s", config.DeployPath, ngxName)
	ret, err := utils.LoadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if ret != ngxConf {
		t.Fatal("ret != ngxConf")
	}

	// wait callback
	time.Sleep(2 * time.Second)
}

func TestDeleteAction(t *testing.T) {
	var err error

	// init backup dir
	if len(config.BackupDir) == 0 {
		t.Fatal("config.BackupDir is null")
	}
	err = os.RemoveAll(config.BackupDir)
	if err != nil {
		t.Fatal(err)
	}

	// exec delete
	ngxConfPerfix := fmt.Sprintf("%s/%s", proWatchPrefix, ngxName)
	opts := &client.DeleteOptions{Recursive: true}
	err = w.client.Delete(ngxConfPerfix, opts)
	if err != nil {
		t.Fatal(err)
	}

	// wait a moment
	time.Sleep(2 * time.Second)
	dirList, err := ioutil.ReadDir(config.BackupDir)
	if err != nil {
		t.Fatal(err)
	}

	isdir, err := utils.IsDir(config.BackupDir)
	if err != nil {
		t.Fatal(err)
	}
	if !isdir {
		t.Fatal("config.BackupDir is null")
	}
	//for _, dir := range dirList {
	//	println(dir.Name())
	//}
	if len(dirList) == 0 {
		t.Fatal("there is no file in the backup dir[%v]", config.BackupDir)
	}

}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	//println("body: ", string(body))

	response, err := Decode(string(body))
	if err != nil {
		log.Fatal(err)
	}

	if response.Code != http.StatusOK {
		log.Fatal("http code doesn't 200")
	}

	if response.Action == "set" || response.Action == "create" {
		if response.MD5 != utils.GetMD5Hash(ngxConf) {
			log.Fatal("http md5 doesn't match")
		}
	}

	if !response.BeforeCmd.Success {
		log.Fatal("execute before cmd failed")
	}
	if !response.AfterCmd.Success {
		log.Fatal("execute after cmd failed")
	}
}

func init() {
	var err error
	w = NewWatcher(cfg)

	// init config
	err = json.Unmarshal([]byte(proConfContent), &config)
	if err != nil {
		xlog.Warn("json.Unmarshal: node:%v, err:%v", prefix, err)
		return
	}
	err = config.checkConfig()
	if err != nil {
		xlog.Warn("checkConfig: node:%v, err:%v", prefix, err)
		return
	}

	// init work dir
	prefix = fmt.Sprintf("%v%v/", w.cfg.Prefix, w.cfg.Hostname)
	w.Lock()
	w.cfg.Prefix = prefix
	w.Unlock()
	proPrefix = fmt.Sprintf("%v/%v", prefix, proName)
	proWatchPrefix = fmt.Sprintf("%v/%v/%v", prefix, proName, EtcdWatchNode)
	proConfPrefix = fmt.Sprintf("%v/%v/%v", prefix, proName, EtcdConfigNode)
	w.client.Delete(prefix, &client.DeleteOptions{Recursive: true})
	w.client.Mkdir(prefix)
	w.client.Mkdir(proPrefix)
	w.client.Mkdir(proWatchPrefix)
	err = w.client.Create(proConfPrefix, []byte(proConfContent))
	if err != nil {
		log.Fatal(err)
	}

	opts := &client.WatcherOptions{Recursive: true}
	go w.client.Watch(proWatchPrefix, opts, w.respProCh, w.exitChan)
	go handleProAction(proPrefix, w, w.exitChan)

	//time.Sleep(2 * time.Second)
	//fmt.Println(prefix)
	//fmt.Println(proPrefix)
	//fmt.Println(proWatchPrefix)
	//fmt.Println(proConfPrefix)
}

func init() {
	go func() {
		http.HandleFunc("/callback", handleCallback)
		err := http.ListenAndServe(":9090", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
