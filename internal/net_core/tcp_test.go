package net_core_test

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/istomyang/call/call-sdk-golang/internal/net_core"
	"github.com/panjf2000/gnet/v2"
)

func Test_Tcp1(t *testing.T) {

	// run echo sever
	go func() {
		s := &tcp_echo_server{}
		s.Run()
	}()

	time.Sleep(time.Second * 3)

	var ctx, cancel = context.WithCancel(context.Background())
	var client = net_core.NewTcpCore(ctx, "127.0.0.1:3000")

	// print msg from server and send msg to server
	go func() {
		for {
			select {
			case data := <-client.Recv():
				log.Println("recv:" + string(data))
			case <-time.Tick(time.Second * 3):
				log.Println("3")
				client.Send([]byte("Hello World!"))
			case <-time.After(time.Minute * 1):
				cancel()
			}
		}
	}()

	client.Run()
}

type tcp_echo_server struct {
}

func (s *tcp_echo_server) Run() error {
	return gnet.Run(s, "tcp://:3000",
		gnet.WithLockOSThread(true),
		gnet.WithLoadBalancing(gnet.RoundRobin),
		gnet.WithTicker(true),
	)
}

// OnBoot implements gnet.EventHandler.
func (s *tcp_echo_server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Println("OnBoot")
	return
}

// OnClose implements gnet.EventHandler.
func (s *tcp_echo_server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	log.Println("OnClose")
	return
}

// OnOpen implements gnet.EventHandler.
func (s *tcp_echo_server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Println("OnOpen")

	return
}

// OnShutdown implements gnet.EventHandler.
func (s *tcp_echo_server) OnShutdown(eng gnet.Engine) {
	log.Println("OnShutdown")

}

// OnTick implements gnet.EventHandler.
func (s *tcp_echo_server) OnTick() (delay time.Duration, action gnet.Action) {
	// log.Println("OnTick")
	return
}

// OnTraffic implements gnet.EventHandler.
func (s *tcp_echo_server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	log.Println("OnTraffic")

	var buf bytes.Buffer
	m, err := c.WriteTo(&buf)
	if err != nil {
		log.Println("OnTraffic:" + err.Error())
		return
	}
	if m == 0 {
		return
	}

	var data = buf.String()
	var new_data = "svr:" + data

	c.Write([]byte(new_data))
	return
}
