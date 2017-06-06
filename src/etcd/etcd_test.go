package etcd

import (
	"testing"
	"time"

	"github.com/coreos/etcd/client"
)

func newTestClient() *EtcdClient {
	c, err := New("127.0.0.1:2379", time.Minute, "", "")
	if err != nil {
		panic(err)
	}
	return c
}

func TestNewEtcdClient(t *testing.T) {
	c := newTestClient()
	defer c.Close()
}

func Test_isErrNoNode(t *testing.T) {
	err := client.Error{}
	err.Code = client.ErrorCodeKeyNotFound
	if !isErrNoNode(err) {
		t.Fatalf("test isErrNoNode failed, %v", err)
	}
	err.Code = client.ErrorCodeNotFile
	if isErrNoNode(err) {
		t.Fatalf("test isErrNoNode failed, %v", err)
	}
}

func Test_isErrNodeExists(t *testing.T) {
	err := client.Error{}
	err.Code = client.ErrorCodeNodeExist
	if !isErrNodeExists(err) {
		t.Fatalf("test isErrNodeExists failed, %v", err)
	}
	err.Code = client.ErrorCodeNotFile
	if isErrNodeExists(err) {
		t.Fatalf("test isErrNodeExists failed, %v", err)
	}
}

func TestMkdir(t *testing.T) {
	c := newTestClient()
	defer c.Close()
	dir := "/ker-unittest/dir"
	err := c.Mkdir(dir)
	if err != nil {
		t.Fatalf("test Mkdir failed, %v", err)
	}
	err = c.Mkdir(dir)
	if err != nil {
		t.Fatalf("test Mkdir failed, %v", err)
	}
}

func TestCreate(t *testing.T) {
	c := newTestClient()
	defer c.Close()
	path := "/ker-unittest/dir/file"
	data := []byte("unittest1")
	err := c.Create(path, data)
	if err != nil {
		t.Fatalf("test Create failed, %v", err)
	}

	err = c.Create(path, data)
	if err == nil {
		t.Fatalf("test Create failed, %v", err)
	}
}

func TestUpdate(t *testing.T) {
	c := newTestClient()
	defer c.Close()
	path := "/ker-unittest/dir/file"
	data := []byte("unittest2")
	err := c.Update(path, data)
	if err != nil {
		t.Fatalf("test Update failed, %v", err)
	}
}

func TestRead(t *testing.T) {
	c := newTestClient()
	defer c.Close()
	path := "/ker-unittest/dir/file"
	b, err := c.Read(path)
	if err != nil {
		t.Fatalf("test read failed, %v", err)
	}
	if string(b) != "unittest2" {
		t.Fatalf("test read failed, not expected data, %s<-->%s", string(b), "unittest2")
	}
}

func TestList(t *testing.T) {
	c := newTestClient()
	defer c.Close()
	path := "/ker-unittest/dir"
	data, err := c.List(path)
	if err != nil {
		t.Fatalf("test list failed, %v", err)
	}
	tmp := []string{"/ker-unittest/dir/file"}
	if data[0] != tmp[0] {
		t.Fatalf("test list failed, not expected data, %v<-->%v", data, tmp)
	}
}

func TestDelete(t *testing.T) {
	c := newTestClient()
	defer c.Close()
	path := "/ker-unittest/dir/file"
	err := c.Delete(path)
	if err != nil {
		t.Fatalf("test delete failed, %v", err)
	}
}

func TestWatch(t *testing.T) {
	c := newTestClient()
	defer c.Close()
	path := "/ker-unittest/dir"
	ch := make(chan string, 1)
	go c.Watch(path, ch)
}
