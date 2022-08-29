package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/colinmarc/hdfs/v2"
	"github.com/colinmarc/hdfs/v2/hadoopconf"
)

type Manager struct {
	cachedClients map[string]*hdfs.Client
}

func NewManager() *Manager {
	return &Manager{
		cachedClients: map[string]*hdfs.Client{},
	}
}

func (m *Manager) getClient(namenode string) (*hdfs.Client, error) {
	if m.cachedClients[namenode] != nil {
		return m.cachedClients[namenode], nil
	}

	if namenode == "" {
		namenode = os.Getenv("HADOOP_NAMENODE")
	}

	conf, err := hadoopconf.LoadFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("Problem loading configuration: %s", err)
	}

	options := hdfs.ClientOptionsFromConf(conf)
	if namenode != "" {
		options.Addresses = strings.Split(namenode, ",")
	}

	if options.Addresses == nil {
		return nil, errors.New("Couldn't find a namenode to connect to. You should specify hdfs://<namenode>:<port> in your paths. Alternatively, set HADOOP_NAMENODE or HADOOP_CONF_DIR in your environment.")
	}

	// if options.KerberosClient != nil {
	// 	options.KerberosClient, err = getKerberosClient()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("Problem with kerberos authentication: %s", err)
	// 	}
	// } else {
	options.User = os.Getenv("HADOOP_USER_NAME")
	if options.User == "" {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("Couldn't determine user: %s", err)
		}

		options.User = u.Username
	}
	// }

	// Set some basic defaults.
	dialFunc := (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 5 * time.Second,
		DualStack: true,
	}).DialContext

	options.NamenodeDialFunc = dialFunc
	options.DatanodeDialFunc = dialFunc

	c, err := hdfs.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to namenode: %s", err)
	}

	m.cachedClients[namenode] = c
	return c, nil
}

func main() {
	manager := NewManager()
	cli, err := manager.getClient("localhost:9000")
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// mkdir dir
	// var mode = 0755 | os.ModeDir
	// if err := cli.Mkdir("/test", mode); err != nil {
	// 	log.Fatal(err)
	// }

	// write into hdfs file
	// writer, err := cli.Create("/test/doc")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// writer.Write([]byte("test text"))
	// writer.Close()

	// read from hdfs file
	file, err := cli.Open("/test/test.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buf := make([]byte, 100)
	if _, err = file.Read(buf); err != nil {
		log.Fatal(err)
	}
	log.Println(string(buf))
}
