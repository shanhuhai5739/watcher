package watcher

import (
	"fmt"
	"strings"

	"etcd"
	"utils/xlog"

	"errors"
	"github.com/coreos/etcd/client"
	"heartbeat"
	"sync"
	"time"
)

var (
	EtcdConfigNode          = "config"
	EtcdWatchNode           = "config.d"
	ErrorEtcdConfigNotFound = errors.New("config doesn't fond of etcd")

	TimeFormat = "2006-01-02_03:04:05"
)

type Watcher struct {
	sync.RWMutex

	cfg       Cfg
	client    *etcd.EtcdClient
	proKey    []string // project's key
	respCh    chan *client.Response
	respProCh chan *client.Response
	exitChan  chan bool
}

func NewWatcher(cfg Cfg) *Watcher {
	cli, err := etcd.New(cfg.Endpoints, cfg.DialTimeout, cfg.Username, cfg.Password)
	if err != nil {
		panic(err)
	}

	w := &Watcher{
		cfg:       cfg,
		client:    cli,
		proKey:    []string{},
		respCh:    make(chan *client.Response),
		respProCh: make(chan *client.Response),
		exitChan:  make(chan bool),
	}
	go w.handleAction()

	return w
}

func (w *Watcher) handleAction() {
	xlog.Debug("handleAction goroutine running")
	for {
		select {
		case resp, ok := <-w.respCh:
			if !ok {
				xlog.Warn("recv from resp chan failed, channel may be closed. node:%v", w.cfg.Prefix)
				goto exit
			}

			switch resp.Action {
			case "get", "update":
			// TODO: noting
			case "create", "set":
				/*
					proPrefix: project's key, like "/watcher/rsyslog"
					proConfdPrefix: project config.d's key, like "/watcher/rsyslog/config.d"
				*/
				proPrefix := trimProPrefix(resp.Node.Key, w.cfg.Prefix)
				proConfdPrefix := fmt.Sprintf("%s/%s", proPrefix, EtcdWatchNode)
				// avoid monitoring multiple project's key
				keyExist := false
				w.Lock()
				for _, key := range w.proKey {
					if key == proPrefix {
						keyExist = true
					}
				}
				w.Unlock()
				if keyExist {
					//xlog.Debug("prefix:%v already watched", proWatchPrefix)
					continue
				}
				w.proKey = append(w.proKey, proPrefix)
				fmt.Println("pro key: ", w.proKey)

				opts := &client.WatcherOptions{Recursive: true}
				go w.client.Watch(proConfdPrefix, opts, w.respProCh, w.exitChan)
				go handleProAction(proPrefix, w, w.exitChan)
			case "delete":
				//prefix := fmt.Sprintf("%s/%s", resp.Node.Key, EtcdWatchNode)
				//xlog.Debug("cannel watch prefix :%v", prefix)
				// TODO: cannel watch goroutine
				// in the etcd that mutil watch a some key is not problems
			}
		case <-w.exitChan:
			goto exit
		}
	}

exit:
	xlog.Debug("handleAction goroutine ending")
}

func (w *Watcher) getConfig(proPrefix string) (prefix string, conf []byte, err error) {
	prefix = fmt.Sprintf("%s/%s", proPrefix, EtcdConfigNode)
	conf, err = w.client.Read(prefix)
	if err != nil {
		return
	}

	return
}

func (w *Watcher) Heartbeat() {
	if len(w.cfg.Heartbeat) == 0 {
		return
	}
	xlog.Debug("Heartbeat goroutine running")
	interval := w.cfg.HeartbeatInterval
	hostname := w.cfg.Hostname
	version := w.cfg.Version
	url := w.cfg.Heartbeat

	timeTicker := time.NewTicker(interval)
	for {
		select {
		case <-timeTicker.C:
			h := &heartbeat.Heartbeat{
				Version:   version,
				Hostname:  hostname,
				Timestamp: time.Now().Unix(),
			}
			err := h.Callback(url)
			if err != nil {
				xlog.Warn("Heartbeat: callback is err, err:%v", err)
				continue
			}
		case <-w.exitChan:
			goto exit
		}
	}
exit:
	timeTicker.Stop()
	xlog.Debug("Heartbeat goroutine ending")
}

func trimProPrefix(s, prefix string) string {
	var str string
	st := strings.TrimPrefix(s, prefix)
	stArr := strings.Split(st, "/")
	str = fmt.Sprintf("%s%s", prefix, stArr[0])

	return str
}

func (w *Watcher) Run() {
	prefix := w.cfg.Prefix
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	prefix = fmt.Sprintf("%v%v/", prefix, w.cfg.Hostname)
	w.Lock()
	w.cfg.Prefix = prefix
	w.Unlock()

	// watch node
	xlog.Debug("watcher prefix %v", prefix)
	opts := &client.WatcherOptions{Recursive: true}
	go w.client.Watch(prefix, opts, w.respCh, w.exitChan)

	// heartbeat
	go w.Heartbeat()
}

func (w *Watcher) Exit() {
	xlog.Debug("watcher ending...")
	close(w.exitChan)
	close(w.respCh)
	close(w.respProCh)
	w.client.Close()
	xlog.Debug("watcher shutdown")
}
