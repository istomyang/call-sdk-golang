package call

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/istomyang/call/call-sdk-golang/internal/net_core"
)

type Request struct {
	ServiceName   string
	ServiceNumber *string
	LB            LoadBalanceAlgo
	WriteOp       bool

	Data []byte
}

type Response struct {
	ServiceName   string
	ServiceNumber string

	Success     bool
	Description string

	Data []byte
}

type message struct {
	Id        string `json:"id"`
	RequestId string `json:"request_id"`

	FromServiceName string  `json:"from_service_name"`
	FromServiceNo   string  `json:"from_service_no"`
	ToServiceName   string  `json:"to_service_name"`
	ToServiceNo     *string `json:"to_service_no"` // force assign

	LBAlgo LoadBalanceAlgo `json:"lb_algo"`

	IsRequest   bool   `json:"is_request"`
	WriteOp     bool   `json:"write_op"`
	Success     bool   `json:"success"`
	Description string `json:"description"`

	Data []byte `json:"data"`
}

func (m *message) ToResponse() *Response {
	return &Response{
		ServiceName:   m.ToServiceName,
		ServiceNumber: *m.ToServiceNo,
		Success:       m.Success,
		Description:   m.Description,
		Data:          m.Data,
	}
}

func (m *Request) ToMessage() message {
	var msgId = getMessageId()
	var requestId = getRequestId()
	return message{
		Id:              msgId,
		RequestId:       requestId,
		FromServiceName: serviceName,
		FromServiceNo:   serviceNo,
		ToServiceName:   m.ServiceName,
		ToServiceNo:     m.ServiceNumber,
		LBAlgo:          m.LB,
		IsRequest:       true,
		WriteOp:         m.WriteOp,
		Success:         false,
		Description:     "",
		Data:            m.Data,
	}
}

type LoadBalanceAlgo = string

const (
	LB_ROUND_ROBIN              LoadBalanceAlgo = "RR"
	LB_Weight_Round_Robin       LoadBalanceAlgo = "WRR"
	LB_Least_Connections        LoadBalanceAlgo = "LC"
	LB_Weight_Least_Connections LoadBalanceAlgo = "WLC"
	LB_Source_IP_Hashing        LoadBalanceAlgo = "SIPH"
	LB_RANDOM                   LoadBalanceAlgo = "Random"
	LB_Response_Time            LoadBalanceAlgo = "RT"
	LB_Power_Of_Two_Choices     LoadBalanceAlgo = "P2C"
)

// ==== Init =====

var (
	ins  runner
	once sync.Once

	serviceName string
	serviceNo   string
)

func Init(ctx context.Context, name, number string, weight uint8, addr string) {
	once.Do(func() {
		var core net_core.NetCore
		{
			ip, _, err := net.SplitHostPort(addr)
			if err != nil {
				panic(err)
			}
			ipAddr := net.ParseIP(ip)
			if ipAddr.IsPrivate() {
				core = net_core.NewUdpCore(ctx, addr)
			} else {
				core = net_core.NewTcpCore(ctx, addr)
			}
		}
		serviceName = name
		serviceNo = number
		ins = newRunner(ctx, core, &registerRequest{
			ServiceName:   name,
			ServiceNumber: number,
			Weight:        weight,
		})
	})
}

func Run() error {
	return ins.Run()
}

// ==== Client ====

func Send(req Request) (res *Response, err error) {
	var a <-chan message
	if a, err = ins.Request(req.ToMessage()); err != nil {
		return
	}
	b := <-a
	res = b.ToResponse()
	return
}

var ErrorTimeOut = errors.New("timeout")

func SendWithTimeout(req Request, timeout time.Duration) (res *Response, err error) {
	var a <-chan message
	if a, err = ins.Request(req.ToMessage()); err != nil {
		return
	}
	select {
	case <-time.After(timeout):
		go func() {
			<-a // use it.
		}()
		err = ErrorTimeOut
		return
	case r := <-a:
		res = r.ToResponse()
		return
	}
}

var ErrorContextDone = errors.New("context done")

func SendWithContext(req Request, ctx context.Context) (res *Response, err error) {
	var a <-chan message
	if a, err = ins.Request(req.ToMessage()); err != nil {
		return
	}
	select {
	case <-ctx.Done():
		go func() {
			<-a // use it.
		}()
		err = ErrorContextDone
		return
	case r := <-a:
		res = r.ToResponse()
		return
	}
}

// ==== Server ====

var onceHandle sync.Once

func Handle(handler func(data []byte) (res []byte, err error)) {
	onceHandle.Do(func() {
		go func() {
			for msg := range ins.Recv() {
				res, err := handler(msg.Data)
				ins.Send(message{
					Id:              getMessageId(),
					RequestId:       msg.RequestId,
					FromServiceName: msg.ToServiceName, // local
					FromServiceNo:   *msg.ToServiceNo,
					ToServiceName:   msg.FromServiceName, // reply to
					ToServiceNo:     &msg.FromServiceNo,
					LBAlgo:          msg.LBAlgo, // invalid
					IsRequest:       false,
					WriteOp:         msg.WriteOp,
					Success:         err != nil,
					Description:     err.Error(),
					Data:            res,
				})
			}
		}()
	})
}
