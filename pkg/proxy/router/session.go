// Copyright 2016 CodisLabs. All Rights Reserved.
// Licensed under the MIT (MIT-LICENSE.txt) license.

package router

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/CodisLabs/codis/pkg/proxy/redis"
	"github.com/CodisLabs/codis/pkg/utils/errors"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/CodisLabs/codis/pkg/utils/sync2/atomic2"
)

type Session struct {
	Conn *redis.Conn

	Ops int64

	CreateUnix int64
	LastOpUnix int64

	auth       string
	authorized bool

	quit   bool
	failed atomic2.Bool

	exit sync.Once
}

func (s *Session) String() string {
	o := &struct {
		Ops        int64  `json:"ops"`
		CreateUnix int64  `json:"create"`
		LastOpUnix int64  `json:"lastop,omitempty"`
		RemoteAddr string `json:"remote"`
	}{
		s.Ops, s.CreateUnix, s.LastOpUnix,
		s.Conn.Sock.RemoteAddr().String(),
	}
	b, _ := json.Marshal(o)
	return string(b)
}

func NewSession(c net.Conn, auth string) *Session {
	return NewSessionSize(c, auth, 1024*32, 1800)
}

func NewSessionSize(c net.Conn, auth string, bufsize int, seconds int) *Session {
	s := &Session{CreateUnix: time.Now().Unix(), auth: auth}
	s.Conn = redis.NewConnSize(c, bufsize)
	s.Conn.ReaderTimeout = time.Second * time.Duration(seconds)
	s.Conn.WriterTimeout = time.Second * 30
	log.Infof("session [%p] create: %s", s, s)
	return s
}

func (s *Session) SetKeepAlivePeriod(period int) error {
	if period == 0 {
		return nil
	}
	if err := s.Conn.SetKeepAlive(true); err != nil {
		return err
	}
	return s.Conn.SetKeepAlivePeriod(time.Second * time.Duration(period))
}

func (s *Session) CloseWithError(err error) {
	s.exit.Do(func() {
		if err != nil {
			log.Infof("session [%p] closed: %s, error: %s", s, s, err)
		} else {
			log.Infof("session [%p] closed: %s, quit", s, s)
		}
		s.Conn.Close()
	})
}

func (s *Session) Start(d Dispatcher, maxPipeline int) {
	tasks := make(chan *Request, maxPipeline)
	var ch = make(chan struct{})

	go func() {
		defer close(ch)
		s.loopWriter(tasks)
	}()

	go func() {
		incrSessions()
		s.loopReader(tasks, d)
		<-ch
		decrSessions()
	}()
}

func (s *Session) loopReader(tasks chan<- *Request, d Dispatcher) (err error) {
	defer func() {
		if err != nil {
			s.CloseWithError(err)
		}
		close(tasks)
	}()
	for !s.quit {
		resp, err := s.Conn.Reader.Decode()
		if err != nil {
			return err
		}
		incrOpTotal()

		r, err := s.handleRequest(resp, d)
		if err != nil {
			incrOpFails()
			return err
		} else {
			tasks <- r
		}
	}
	return nil
}

func (s *Session) loopWriter(tasks <-chan *Request) (err error) {
	defer func() {
		s.CloseWithError(err)
		for _ = range tasks {
			incrOpFails()
		}
	}()
	p := &FlushPolicy{
		Encoder:     s.Conn.Writer,
		MaxBuffered: 32,
		MaxInterval: 300,
	}
	for r := range tasks {
		resp, err := s.handleResponse(r)
		if err != nil {
			incrOpFails()
			return err
		}
		if err := p.Encode(resp, len(tasks) == 0); err != nil {
			incrOpFails()
			return err
		}
	}
	return nil
}

var ErrRespIsRequired = errors.New("resp is required")

func (s *Session) handleResponse(r *Request) (*redis.Resp, error) {
	r.Wait.Wait()
	if r.Coalesce != nil {
		if err := r.Coalesce(); err != nil {
			return nil, err
		}
	}
	resp, err := r.Response.Resp, r.Response.Err
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, ErrRespIsRequired
	}
	incrOpStats(r.OpStr, microseconds()-r.Start)
	return resp, nil
}

func (s *Session) handleRequest(resp *redis.Resp, d Dispatcher) (*Request, error) {
	opstr, err := getOpStr(resp)
	if err != nil {
		return nil, err
	}
	if isNotAllowed(opstr) {
		return nil, errors.New(fmt.Sprintf("command <%s> is not allowed", opstr))
	}

	usnow := microseconds()
	s.LastOpUnix = usnow / 1e6
	s.Ops++

	r := &Request{
		OpStr:  opstr,
		Start:  usnow,
		Resp:   resp,
		Wait:   &sync.WaitGroup{},
		Failed: &s.failed,
	}

	if opstr == "QUIT" {
		return s.handleQuit(r)
	}
	if opstr == "AUTH" {
		return s.handleAuth(r)
	}

	if !s.authorized {
		if s.auth != "" {
			r.Response.Resp = redis.NewError([]byte("NOAUTH Authentication required."))
			return r, nil
		}
		s.authorized = true
	}

	switch opstr {
	case "SELECT":
		return s.handleSelect(r)
	case "PING":
		return s.handlePing(r)
	case "MGET":
		return s.handleRequestMGet(r, d)
	case "MSET":
		return s.handleRequestMSet(r, d)
	case "DEL":
		return s.handleRequestMDel(r, d)
	}
	return r, d.Dispatch(r)
}

func (s *Session) handleQuit(r *Request) (*Request, error) {
	s.quit = true
	r.Response.Resp = redis.NewString([]byte("OK"))
	return r, nil
}

func (s *Session) handleAuth(r *Request) (*Request, error) {
	if len(r.Resp.Array) != 2 {
		r.Response.Resp = redis.NewError([]byte("ERR wrong number of arguments for 'AUTH' command"))
		return r, nil
	}
	if s.auth == "" {
		r.Response.Resp = redis.NewError([]byte("ERR Client sent AUTH, but no password is set"))
		return r, nil
	}
	if s.auth != string(r.Resp.Array[1].Value) {
		s.authorized = false
		r.Response.Resp = redis.NewError([]byte("ERR invalid password"))
		return r, nil
	} else {
		s.authorized = true
		r.Response.Resp = redis.NewString([]byte("OK"))
		return r, nil
	}
}

func (s *Session) handleSelect(r *Request) (*Request, error) {
	if len(r.Resp.Array) != 2 {
		r.Response.Resp = redis.NewError([]byte("ERR wrong number of arguments for 'SELECT' command"))
		return r, nil
	}
	if db, err := strconv.Atoi(string(r.Resp.Array[1].Value)); err != nil {
		r.Response.Resp = redis.NewError([]byte("ERR invalid DB index"))
		return r, nil
	} else if db != 0 {
		r.Response.Resp = redis.NewError([]byte("ERR invalid DB index, only accept DB 0"))
		return r, nil
	} else {
		r.Response.Resp = redis.NewString([]byte("OK"))
		return r, nil
	}
}

func (s *Session) handlePing(r *Request) (*Request, error) {
	if len(r.Resp.Array) != 1 {
		r.Response.Resp = redis.NewError([]byte("ERR wrong number of arguments for 'PING' command"))
		return r, nil
	}
	r.Response.Resp = redis.NewString([]byte("PONG"))
	return r, nil
}

func (s *Session) handleRequestMGet(r *Request, d Dispatcher) (*Request, error) {
	nkeys := len(r.Resp.Array) - 1
	if nkeys <= 1 {
		return r, d.Dispatch(r)
	}
	var sub = make([]*Request, nkeys)
	for i := 0; i < len(sub); i++ {
		sub[i] = &Request{
			OpStr: r.OpStr,
			Start: r.Start,
			Resp: redis.NewArray([]*redis.Resp{
				r.Resp.Array[0],
				r.Resp.Array[i+1],
			}),
			Wait:   r.Wait,
			Failed: r.Failed,
		}
		if err := d.Dispatch(sub[i]); err != nil {
			return nil, err
		}
	}
	r.Coalesce = func() error {
		var array = make([]*redis.Resp, len(sub))
		for i, x := range sub {
			if err := x.Response.Err; err != nil {
				return err
			}
			resp := x.Response.Resp
			if resp == nil {
				return ErrRespIsRequired
			}
			if !resp.IsArray() || len(resp.Array) != 1 {
				return errors.New(fmt.Sprintf("bad mget resp: %s array.len = %d", resp.Type, len(resp.Array)))
			}
			array[i] = resp.Array[0]
		}
		r.Response.Resp = redis.NewArray(array)
		return nil
	}
	return r, nil
}

func (s *Session) handleRequestMSet(r *Request, d Dispatcher) (*Request, error) {
	nblks := len(r.Resp.Array) - 1
	if nblks <= 2 {
		return r, d.Dispatch(r)
	}
	if nblks%2 != 0 {
		r.Response.Resp = redis.NewError([]byte("ERR wrong number of arguments for 'MSET' command"))
		return r, nil
	}
	var sub = make([]*Request, nblks/2)
	for i := 0; i < len(sub); i++ {
		sub[i] = &Request{
			OpStr: r.OpStr,
			Start: r.Start,
			Resp: redis.NewArray([]*redis.Resp{
				r.Resp.Array[0],
				r.Resp.Array[i*2+1],
				r.Resp.Array[i*2+2],
			}),
			Wait:   r.Wait,
			Failed: r.Failed,
		}
		if err := d.Dispatch(sub[i]); err != nil {
			return nil, err
		}
	}
	r.Coalesce = func() error {
		for _, x := range sub {
			if err := x.Response.Err; err != nil {
				return err
			}
			resp := x.Response.Resp
			if resp == nil {
				return ErrRespIsRequired
			}
			if !resp.IsString() {
				return errors.New(fmt.Sprintf("bad mset resp: %s value.len = %d", resp.Type, len(resp.Value)))
			}
			r.Response.Resp = resp
		}
		return nil
	}
	return r, nil
}

func (s *Session) handleRequestMDel(r *Request, d Dispatcher) (*Request, error) {
	nkeys := len(r.Resp.Array) - 1
	if nkeys <= 1 {
		return r, d.Dispatch(r)
	}
	var sub = make([]*Request, nkeys)
	for i := 0; i < len(sub); i++ {
		sub[i] = &Request{
			OpStr: r.OpStr,
			Start: r.Start,
			Resp: redis.NewArray([]*redis.Resp{
				r.Resp.Array[0],
				r.Resp.Array[i+1],
			}),
			Wait:   r.Wait,
			Failed: r.Failed,
		}
		if err := d.Dispatch(sub[i]); err != nil {
			return nil, err
		}
	}
	r.Coalesce = func() error {
		var n int
		for _, x := range sub {
			if err := x.Response.Err; err != nil {
				return err
			}
			resp := x.Response.Resp
			if resp == nil {
				return ErrRespIsRequired
			}
			if !resp.IsInt() || len(resp.Value) != 1 {
				return errors.New(fmt.Sprintf("bad mdel resp: %s value.len = %d", resp.Type, len(resp.Value)))
			}
			if resp.Value[0] != '0' {
				n++
			}
		}
		r.Response.Resp = redis.NewInt([]byte(strconv.Itoa(n)))
		return nil
	}
	return r, nil
}

func microseconds() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}
