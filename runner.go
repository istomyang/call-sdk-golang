package call

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/istomyang/call/call-sdk-golang/internal/net_core"
)

type runner interface {
	Run() (err error)
	Request(req message) (<-chan message, error)
	Send(req message)
	Recv() <-chan message
}

type runnerImpl struct {
	ctx  context.Context
	core net_core.NetCore

	mut     sync.RWMutex
	futures []*future

	send chan message
	recv chan message

	reg *registerRequest
}

func newRunner(ctx context.Context, core net_core.NetCore, reg *registerRequest) runner {
	return &runnerImpl{
		ctx:  ctx,
		core: core,
		mut:  sync.RWMutex{},
		recv: make(chan message),
		send: make(chan message),
		reg:  reg,
	}
}

var _ runner = &runnerImpl{}

func (r *runnerImpl) Request(req message) (res <-chan message, err error) {
	var result = make(chan message)
	var future = future{
		msg_id: req.Id,
		result: result,
	}

	r.mut.Lock()
	r.futures = append(r.futures, &future)
	r.mut.Unlock()

	r.core.Send(req.Data)

	res = result
	return
}

// Recv implements runner.
func (r *runnerImpl) Recv() <-chan message {
	return r.recv
}

// Send implements runner.
func (r *runnerImpl) Send(req message) {
	r.send <- req
}

func (r *runnerImpl) Run() (err error) {
	// register
	{
		r.core.Send(encode(r.reg))
		b := <-r.core.Recv() // blocking
		var res registerResponse
		if err := decode(b, &res); err != nil {
			return err
		}
		if !res.Success {
			return errors.New(res.Description)
		}
	}

	// get results.
	go func() {
		var ch = r.core.Recv()
		for b := range ch {
			var msg message
			if err := decode(b, &msg); err != nil {
				continue
			}

			// 1. request from remote service.
			if msg.IsRequest {
				r.recv <- msg
				continue
			}

			// 2. response for request.
			r.mut.RLock()
			for _, f := range r.futures {
				if f.msg_id != msg.Id {
					continue
				}
				// one thread visit one future, write op is ok.
				f.result <- msg
				f.flag_deleted = true
				break
			}
			r.mut.RUnlock()
		}
	}()

	// handle send to remote
	go func() {
		for req := range r.send {
			r.core.Send(encode(req))
		}
	}()

	// clean futures.
	go func() {
		var t = time.NewTicker(time.Second * 10) // further estimate.
		defer t.Stop()
		for {
			select {
			case <-t.C:
				r.mut.Lock()
				for i, f := range r.futures {
					if f.flag_deleted {
						r.futures[i] = r.futures[len(r.futures)-1]
						r.futures = r.futures[:len(r.futures)-1]
					}
				}
				r.mut.Unlock()
			case <-r.ctx.Done():
				return
			}
		}
	}()

	return r.core.Run()
}

type future struct {
	msg_id string
	result chan message

	flag_deleted bool
}

type registerRequest struct {
	ServiceName   string `json:"service_name"`
	ServiceNumber string `json:"service_no"`
	Weight        uint8  `json:"weight"`
}

type registerResponse struct {
	Success     bool   `json:"success"`
	Description string `json:"description"`
}
