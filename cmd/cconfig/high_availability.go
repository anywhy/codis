package main

import (
	"time"
	"github.com/CodisLabs/codis/pkg/models"
	"github.com/juju/errors"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/CodisLabs/codis/pkg/utils"
	"github.com/docopt/docopt-go"
	"strconv"
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

	saveMap = make(map[int]string)
)

func cmdCodisHA(argv []string) (err error) {
	usage := `usage: codis-config codis-ha [--interval=<seconds>]

options:
	--interval=<seconds>	set monitor ha interval [default: 3]
`

	args, err := docopt.Parse(usage, argv, true, "", false)
	if err != nil {
		log.ErrorErrorf(err, "parse args failed")
		return err
	}
	log.Debugf("parse args = {%+v}", args)

	interval := 3
	if s, ok := args["--interval"].(string); ok && s != "" {
		n, err := strconv.Atoi(s);
		if (err != nil) {
			log.Error(err)
		}
		if n <= 0 {
			log.Panicf("option --interval = %d", n)
		}
		interval = n
	}

	runCodisHA(interval)

	return nil
}

func runCodisHA(interval int) {
	for {
		groups, err := runGetServerGroups()
		if err != nil {
			log.Error(err)
			return
		}

		CheckAliveAndPromote(groups)
		CheckOfflineAndPromoteSlave(groups)
		time.Sleep(time.Duration(interval) * time.Second)
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
	group, err := runGetServerGroup(master.GroupId)
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

		if save, err := utils.GetSaveInfo(slave.Addr, globalEnv.Password()); err == nil {
			saveMap[slave.GroupId] = save
			log.Infof("cache group_id:%s, slave(addr:%s) strategy:%s", strconv.Itoa(slave.GroupId), slave.Addr, save)
		}

		log.Infof("try promote to master %+v", slave)
		err = runPromoteServerToMaster(slave.GroupId, slave.Addr)
		if err != nil {
			log.Errorf("do promote %v failed %v", slave, errors.ErrorStack(err))
			return err
		}
		log.Infof("try promote to master success %+v", slave)
		log.Infof("make master save \"\", group_id:%s, master:%s", strconv.Itoa(slave.GroupId), slave.Addr)
		err = utils.SetSaveInfo(slave.Addr, globalEnv.Password(), "");
		if err != nil {
			log.Errorf("do make master save \"\" %v failed %v", slave, errors.ErrorStack(err))
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
	if (err == nil) {
		save := saveMap[s.GroupId]
		log.Infof("add slave(addr:%s, save:%s) to group_id:%s, ", s.Addr, save, strconv.Itoa(s.GroupId))
		err = utils.SetSaveInfo(s.Addr, globalEnv.Password(), save);
	}

	if (err != nil) {
		log.Errorf("do reusing slave %v failed %v", s, errors.ErrorStack(err))
	}
}

//ping codis-server find crashed codis-server
func CheckAliveAndPromote(groups []models.ServerGroup) ([]models.Server, error) {
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
func CheckOfflineAndPromoteSlave(groups []models.ServerGroup) ([]models.Server, error) {
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
