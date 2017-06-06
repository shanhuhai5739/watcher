package conf

import (
	"sync"

	"flag"
	"github.com/Unknwon/goconfig"
)

var (
	config    *goconfig.ConfigFile
	lock      sync.RWMutex
	remoteUrl string
)

func InitConf(filename string) (err error) {
	config, err = goconfig.LoadConfigFile(filename)
	if err != nil {
		return
	}
	return
}

func Get(sect, key string) (val string, err error) {
	lock.RLock()
	defer lock.RUnlock()

	val, err = config.GetValue(sect, key)
	if err != nil {
		return
	}
	return
}

func Int(sect, key string) (val int, err error) {
	lock.RLock()
	defer lock.RUnlock()

	val, err = config.Int(sect, key)
	if err != nil {
		return
	}
	return
}

func Bool(sect, key string) (val bool, err error) {
	lock.RLock()
	defer lock.RUnlock()

	val, err = config.Bool(sect, key)
	if err != nil {
		return
	}
	return
}

func GetSect(sect string) (val map[string]string, err error) {
	lock.RLock()
	defer lock.RUnlock()

	val, err = config.GetSection(sect)
	if err != nil {
		return
	}
	return
}

func init() {
	var err error
	confname := flag.String("c", "config/scm_config.ini", "-c /path/to/config")
	config, err = goconfig.LoadConfigFile(*confname)
	if err != nil {
		panic(err)
	}
}
