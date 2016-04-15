package main

import (
	"time"
	"github.com/CodisLabs/codis/pkg/models"
	"github.com/juju/errors"
	"github.com/CodisLabs/codis/pkg/utils/log"
)

type aliveCheckerFactory func(addr string, defaultTimeout time.Duration) AliveChecker

var (
	acf aliveCheckerFactory = func(addr string, timeout time.Duration) AliveChecker {
		return &redisChecker{
			addr:           addr,
			defaultTimeout: timeout,
			passwd: globalEnv.Password(),
		}
	}
)

func StartHA() {
	for {
		groups, err := models.ServerGroups(safeZkConn, globalEnv.ProductName())
		if err != nil {
			log.Error(err)
			return
		}

		CheckAliveAndPromote(groups)
		CheckOfflineAndPromoteSlave(groups)
		time.Sleep(10 * time.Second)
	}
}

func PingServer(checker AliveChecker, errCtx interface{}, errCh chan <- interface{}) {
	err := checker.CheckAlive()
	log.Debugf("check %+v, result:%v, errCtx:%+v", checker, err, errCtx)
	if err != nil {
		errCh <- errCtx
		return
	}
	errCh <- nil
}

func verifyAndUpServer(checker AliveChecker, errCtx interface{}) {
	errCh := make(chan interface{}, 100)

	go PingServer(checker, errCtx, errCh)

	s := <-errCh

	if s == nil {
		//alive
		handleAddServer(errCtx.(*models.Server))
	}

}

func getSlave(master *models.Server) (*models.Server, error) {
	group, err := models.GetGroup(safeZkConn, globalEnv.ProductName(), master.GroupId)
	if err != nil {
		return nil, errors.Trace(err)
	}

	for _, s := range group.Servers {
		if s.Type == models.SERVER_TYPE_SLAVE {
			return s, nil
		}
	}

	return nil, errors.Errorf("can not find any slave in this group: %v", group)
}

func handleCrashedServer(s *models.Server) error {
	switch s.Type {
	case models.SERVER_TYPE_MASTER:
		//get slave and do promote
		slave, err := getSlave(s)
		if err != nil {
			log.Warn(errors.ErrorStack(err))
			return err
		}

		log.Infof("try promote %+v", slave)
		err = runPromoteServerToMaster(slave.GroupId, slave.Addr)
		if err != nil {
			log.Errorf("do promote %v failed %v", slave, errors.ErrorStack(err))
			return err
		}
	case models.SERVER_TYPE_SLAVE:
		log.Errorf("slave is down: %+v", s)
	case models.SERVER_TYPE_OFFLINE:
	//no need to handle it
	default:
		log.Errorf("unkonwn type %+v", s)
	}

	return nil
}

func handleAddServer(s *models.Server) {
	s.Type = models.SERVER_TYPE_SLAVE
	log.Infof("try reusing slave %+v", s)
	err := runAddServerToGroup(s.GroupId, s.Addr, s.Type)
	log.Errorf("do reusing slave %v failed %v", s, errors.ErrorStack(err))
}

//ping codis-server find crashed codis-server
func CheckAliveAndPromote(groups []*models.ServerGroup) ([]models.Server, error) {
	errCh := make(chan interface{}, 100)
	var serverCnt int
	for _, group := range groups {
		//each group
		for _, s := range group.Servers {
			//each server
			serverCnt++
			rc := acf(s.Addr, 5 * time.Second)
			news := s
			go PingServer(rc, news, errCh)
		}
	}

	//get result
	var crashedServer []models.Server
	for i := 0; i < serverCnt; i++ {
		s := <-errCh
		if s == nil {
			//alive
			continue
		}

		log.Warnf("server maybe crashed %+v", s)
		crashedServer = append(crashedServer, *s.(*models.Server))

		err := handleCrashedServer(s.(*models.Server))
		if err != nil {
			return crashedServer, err
		}
	}

	return crashedServer, nil
}

//ping codis-server find node up with type offine
func CheckOfflineAndPromoteSlave(groups []*models.ServerGroup) ([]models.Server, error) {
	for _, group := range groups {
		//each group
		for _, s := range group.Servers {
			//each server
			rc := acf(s.Addr, 5 * time.Second)
			news := s
			if (s.Type == models.SERVER_TYPE_OFFLINE) {
				verifyAndUpServer(rc, news)
			}
		}
	}
	return nil, nil
}
