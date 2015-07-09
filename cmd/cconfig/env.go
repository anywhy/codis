// Copyright 2014 Wandoujia Inc. All Rights Reserved.
// Licensed under the MIT (MIT-LICENSE.txt) license.

package main

import (
	"os"
	"strings"

	"github.com/c4pt0r/cfg"
	"github.com/juju/errors"
	log "github.com/ngaut/logging"

	"github.com/wandoulabs/zkhelper"
)

type Env interface {
	ProductName() string
	Password() string
	DashboardAddr() string
	NewZkConn() (zkhelper.Conn, error)
}

type CodisEnv struct {
	zkAddr        string
	passwd        string
	dashboardAddr string
	productName   string
	provider      string
}

func LoadCodisEnv(cfg *cfg.Cfg) Env {
	if cfg == nil {
		log.Fatal("config is nil")
	}

	productName, err := cfg.ReadString("product", "test")
	if err != nil {
		log.Fatal(err)
	}

	zkAddr, err := cfg.ReadString("zk", "localhost:2181")
	if err != nil {
		log.Fatal(err)
	}

	hostname, _ := os.Hostname()
	dashboardAddr, err := cfg.ReadString("dashboard_addr", hostname+":18087")
	if err != nil {
		log.Fatal(err)
	}

	provider, err := cfg.ReadString("coordinator", "zookeeper")
	if err != nil {
		log.Fatal(err)
	}

	passwd, _ := cfg.ReadString("password", "")

	return &CodisEnv{
		zkAddr:        zkAddr,
		passwd:        passwd,
		dashboardAddr: dashboardAddr,
		productName:   productName,
		provider:      provider,
	}
}

func (e *CodisEnv) ProductName() string {
	return e.productName
}

func (e *CodisEnv) Password() string {
	return e.passwd
}

func (e *CodisEnv) DashboardAddr() string {
	return e.dashboardAddr
}

func (e *CodisEnv) NewZkConn() (zkhelper.Conn, error) {
	switch e.provider {
	case "zookeeper":
		return zkhelper.ConnectToZk(e.zkAddr)
	case "etcd":
		addr := strings.TrimSpace(e.zkAddr)
		if !strings.HasPrefix(addr, "http://") {
			addr = "http://" + addr
		}
		return zkhelper.NewEtcdConn(addr)
	}
	return nil, errors.Errorf("need coordinator in config file, %+v", e)
}
