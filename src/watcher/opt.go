package watcher

import (
	"flag"
	"fmt"

	"os"
	"time"
	"utils/conf"
)

var (
	Prefix string
)

type Cfg struct {
	Endpoints   string
	DialTimeout time.Duration
	Hostname    string
	Username    string
	Password    string

	Heartbeat         string
	HeartbeatInterval time.Duration
	Prefix            string
	Force             bool
	Version           string
}

func NewCfg(version string) Cfg {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	// local
	localPrefix, err := conf.Get("local", "prefix")
	checkArg("local.prefix", localPrefix, err)
	localForce, err := conf.Bool("local", "force")
	checkArg("local.force", localForce, err)

	// etcd
	etcdEndpoints, err := conf.Get("etcd", "endpoints")
	checkArg("etcd.endpoints", etcdEndpoints, err)
	etcdTimeout, err := conf.Int("etcd", "timeout")
	checkArg("etcd.timeout", etcdTimeout, err)
	etcdUsername, err := conf.Get("etcd", "username")
	//checkArg("etcd.username", etcdUsername, err)
	etcdPassword, err := conf.Get("etcd", "password")
	//checkArg("etcd.password", etcdPassword, err)

	// heartbeat
	heartbeatDomain, err := conf.Get("heartbeat", "domain")
	checkArg("heartbeat.domain", heartbeatDomain, err)
	heartbeatInterval, err := conf.Int("heartbeat", "interval")
	checkArg("heartbeat.interval", heartbeatInterval, err)

	return Cfg{
		Endpoints:         etcdEndpoints,
		DialTimeout:       time.Duration(etcdTimeout) * time.Second,
		Hostname:          hostname,
		Username:          etcdUsername,
		Password:          etcdPassword,
		Heartbeat:         heartbeatDomain,
		HeartbeatInterval: time.Duration(heartbeatInterval) * time.Second,
		Prefix:            localPrefix,
		Force:             localForce,
		Version:           version,
	}
}

func checkArg(name string, arg interface{}, err error) {
	if err != nil {
		panic(err)
	}
	switch t := arg.(type) {
	case string:
		if len(t) == 0 {
			err = fmt.Errorf("Cfg: %v arg is null", name)
			panic(err)
		}
	case int:
		if t <= 0 {
			err = fmt.Errorf("Cfg: %v arg can't <= 0", name)
			panic(err)
		}
	case bool:
	default:
		err = fmt.Errorf("unexpected type %T", t)
		panic(err)
	}

}

func (cfg *Cfg) checkCfg() (err error) {
	return
}

type Config struct {
	DeployPath string `json:"deployPath"`
	BackupDir  string `json:"backupDir"`
	BeforeCmd  string `json:"beforeCmd"`
	AfterCmd   string `json:"afterCmd"`
	Callback   string `json:"callback"`
}

func (c *Config) checkConfig() (err error) {
	if len(c.DeployPath) == 0 {
		return fmt.Errorf("DeployPath argument is null")
	}
	return
}

func init() {
	flag.StringVar(&Prefix, "prefix", "", "key path prefix")
}
