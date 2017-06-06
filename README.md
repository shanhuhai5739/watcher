# Watcher

## 介绍
watcher是由go语言实现的配置文件分发，基于etcd做配置文件事件的触发。

    1. 支持多项目配置文件发布
    2. 支持配置文件删除时备份
    3. 支持异步回调response信息
    4. 支持心跳检测

## 玩转watcher

### watcher在etcd中的结构
```
/watcher/web01/a.com/config
/watcher/web01/a.com/config.d/

前缀：/watcher
主机：/watcher/web01/
项目：a./watcher/web01/a.com
分发策略配置：/watcher/web01/a.com/config
项目配置文件：/watcher/web01/a.com/config.d/
```


### 编译运行
```
./build watcher
```
交叉编译
<pre>
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build    # linux平台
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build  # windows平台
</pre>

运行
```
./bin/watcher
```


## 接入watcher
场景：一台web01主机需发布a.com和b.com的nginx的配置文件
<pre>
# 发布a.com的发布策略配置文件
./bin/etcdctl set /watcher/coralmac.local/a.com/config '{"deployPath": "/tmp/watcher/a.com/", "backupDir":"/tmp/backup/a.com", "beforeCmd":"echo before", "afterCmd":"echo after", "callback": ""}'
# 发布a.com的nginx配置文件
./bin/etcdctl set /watcher/coralmac.local/a.com/config.d/ngx.conf 'aaaaaaaa'


# 发布b.com的发布策略配置文件
./bin/etcdctl set /watcher/coralmac.local/b.com/config '{"deployPath": "/tmp/watcher/b.com/", "backupDir":"/tmp/backup/b.com", "beforeCmd":"echo before", "afterCmd":"echo after", "callback": ""}'
# 发布b.com的nginx配置文件
./bin/etcdctl set /watcher/coralmac.local/b.com/config.d/ngx.conf 'bbbbbbbbb'
</pre>
写入到etcd后，watcher自动触发，将配置发布到服务器的deployPath路径下


### 发布策略配置说明
```
{"deployPath": "/tmp/watcher", "backupDir":"/tmp/backup", "beforeCmd":"echo before", "afterCmd":"echo after", "callback": "http://www.a.com/callback"}
deployPath: 部署目录
backupDir:  配置备份目录，如果该目录为空则直接删除配置文件，如果不为空则进行配置文件备份到该目录
beforeCmd:  配置发布之前执行的操作
afterCmd:   配置发布之后执行的操作
callback:   项目异步回调的地址，用于提交发布的结果
```


## watcher的运维

### watcher本地配置文件
```
$ cat config/scm_config.ini
[local]                           # watcher相关
prefix = /watcher                 # etcd中的前缀
force = true                      # 是否强制，用于watcher重启后强制同步所有配置

[etcd]                            # etcd相关
endpoints = localhost:2379
timeout = 5
username =
password =

[logs]                            # 日志相关
name = watcher
path = ./logs/
filename = watcher
level = debug

[heartbeat]
domain = http://127.0.0.1:9091    # 心跳配置
interval = 30                     # 心跳提交的间隔时间，以秒为单位
```








