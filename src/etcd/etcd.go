package etcd

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/client"
	//client "github.com/coreos/etcd/clientv3"
	"utils/xlog"
)

var ErrClosedEtcdClient = errors.New("use of closed etcd client")

type EtcdClient struct {
	sync.Mutex
	kapi client.KeysAPI

	closed  bool
	timeout time.Duration
}

func New(addr string, timeout time.Duration, username, passwd string) (*EtcdClient, error) {
	endpoints := strings.Split(addr, ",")
	for i, s := range endpoints {
		if s != "" && !strings.HasPrefix(s, "http://") {
			endpoints[i] = "http://" + s
		}
	}
	config := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		Username:                username,
		Password:                passwd,
		HeaderTimeoutPerRequest: time.Second * 3,
	}
	c, err := client.New(config)
	if err != nil {
		return nil, err
	}
	return &EtcdClient{
		kapi: client.NewKeysAPI(c), timeout: timeout,
	}, nil
}

func (c *EtcdClient) Close() error {
	c.Lock()
	defer c.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return nil
}

func (c *EtcdClient) contextWithTimeout() (context.Context, context.CancelFunc) {
	if c.timeout == 0 {
		return context.Background(), func() {}
	} else {
		return context.WithTimeout(context.Background(), c.timeout)
	}
}

func isErrNoNode(err error) bool {
	if err != nil {
		if e, ok := err.(client.Error); ok {
			return e.Code == client.ErrorCodeKeyNotFound
		}
	}
	return false
}

func isErrNodeExists(err error) bool {
	if err != nil {
		if e, ok := err.(client.Error); ok {
			return e.Code == client.ErrorCodeNodeExist
		}
	}
	return false
}

func (c *EtcdClient) Mkdir(dir string) error {
	c.Lock()
	defer c.Unlock()
	if c.closed {
		return ErrClosedEtcdClient
	}
	return c.mkdir(dir)
}

func (c *EtcdClient) mkdir(dir string) error {
	if dir == "" || dir == "/" {
		return nil
	}
	cntx, canceller := c.contextWithTimeout()
	defer canceller()
	_, err := c.kapi.Set(cntx, dir, "", &client.SetOptions{Dir: true, PrevExist: client.PrevNoExist})
	if err != nil {
		if isErrNodeExists(err) {
			return nil
		}
		return err
	}
	return nil
}

func (c *EtcdClient) Create(path string, data []byte) error {
	c.Lock()
	//defer c.Unlock()
	if c.closed {
		return ErrClosedEtcdClient
	}
	c.Unlock()
	cntx, canceller := c.contextWithTimeout()
	defer canceller()
	_, err := c.kapi.Set(cntx, path, string(data), &client.SetOptions{PrevExist: client.PrevNoExist})
	if err != nil {
		xlog.Debug("etcd create node %s failed: %s", path, err)
		return err
	}
	xlog.Debug("etcd create node %s OK", path)
	return nil
}

func (c *EtcdClient) Update(path string, data []byte) error {
	c.Lock()
	//defer c.Unlock()
	if c.closed {
		return ErrClosedEtcdClient
	}
	c.Unlock()
	cntx, canceller := c.contextWithTimeout()
	defer canceller()
	_, err := c.kapi.Set(cntx, path, string(data), &client.SetOptions{PrevExist: client.PrevIgnore})
	if err != nil {
		xlog.Debug("etcd update node %s failed: %s", path, err)
		return err
	}
	xlog.Debug("etcd update node %s OK", path)
	return nil
}

func (c *EtcdClient) Delete(path string, opts *client.DeleteOptions) error {
	c.Lock()
	//defer c.Unlock()
	if c.closed {
		return ErrClosedEtcdClient
	}
	c.Unlock()
	cntx, canceller := c.contextWithTimeout()
	defer canceller()
	_, err := c.kapi.Delete(cntx, path, opts)
	if err != nil && !isErrNoNode(err) {
		xlog.Debug("etcd delete node %s failed: %s", path, err)
		return err
	}
	xlog.Debug("etcd delete node %s OK", path)
	return nil
}

func (c *EtcdClient) Read(path string) ([]byte, error) {
	c.Lock()
	//defer c.Unlock()
	if c.closed {
		return nil, ErrClosedEtcdClient
	}
	c.Unlock()
	cntx, canceller := c.contextWithTimeout()
	defer canceller()
	xlog.Debug("etcd read node %s", path)
	r, err := c.kapi.Get(cntx, path, nil)
	if err != nil && !isErrNoNode(err) {
		return nil, err
	} else if r == nil || r.Node.Dir {
		return nil, nil
	} else {
		return []byte(r.Node.Value), nil
	}
}

func (c *EtcdClient) List(path string) ([]string, error) {
	c.Lock()
	//defer c.Unlock()
	if c.closed {
		return nil, ErrClosedEtcdClient
	}
	c.Unlock()

	cntx, canceller := c.contextWithTimeout()
	defer canceller()
	xlog.Debug("etcd list node %s", path)
	r, err := c.kapi.Get(cntx, path, nil)
	if err != nil && !isErrNoNode(err) {
		return nil, err
	} else if r == nil || !r.Node.Dir {
		return nil, nil
	} else {
		var files []string
		for _, node := range r.Node.Nodes {
			files = append(files, node.Key)
		}
		return files, nil
	}
}

func (c *EtcdClient) Watch(path string, opts *client.WatcherOptions, respCh chan *client.Response, exitCh chan bool) {
	c.Lock()
	if c.closed {
		panic(ErrClosedEtcdClient)
	}
	c.Unlock()

	watcher := c.kapi.Watcher(path, opts)
	ctx, cancel := context.WithCancel(context.Background())
	cancelRoutine := make(chan bool)
	defer close(cancelRoutine)

	go func() {
		select {
		case <-exitCh:
			cancel()
		case <-cancelRoutine:
			return
		}
	}()

	for {
		res, err := watcher.Next(ctx)
		if err != nil {
			xlog.Fatal(err.Error())
			return
		}
		if !c.closed {
			respCh <- res
		}
	}
}
