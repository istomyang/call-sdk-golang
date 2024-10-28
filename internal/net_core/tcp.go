package net_core

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type tcpCore struct {
	ctx context.Context

	addr string

	recv chan []byte
	send chan []byte
}

func NewTcpCore(ctx context.Context, addr string) NetCore {
	return &tcpCore{
		ctx:  ctx,
		addr: addr,
		recv: make(chan []byte),
		send: make(chan []byte),
	}
}

var _ NetCore = &tcpCore{}

// Recv implements NetCore.
func (n *tcpCore) Recv() <-chan []byte {
	return n.recv
}

// Send implements NetCore.
func (n *tcpCore) Send(data []byte) {
	n.send <- data
}

// Run implements NetCore.
func (n *tcpCore) Run() (err error) {
	var client *gnet.Client
	if client, err = gnet.NewClient(n, gnet.WithLockOSThread(true),
		gnet.WithTicker(true),
		gnet.WithLoadBalancing(gnet.RoundRobin),
		gnet.WithMulticore(true)); err != nil {
		return
	}
	client.Start()

	var conn gnet.Conn
	if conn, err = client.DialContext("tcp", n.addr, n.ctx); err != nil {
		return
	}
	defer conn.Close()

	for {
		select {
		case <-n.ctx.Done():
			n.close()
			return
		case data := <-n.send:
			conn.Write(data)
		}
	}
}

func (n *tcpCore) close() {
	close(n.send)
	close(n.recv)
}

var _ gnet.EventHandler = &tcpCore{}

// OnBoot implements gnet.EventHandler.
func (n *tcpCore) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Println("client: OnBoot")
	return
}

// OnClose implements gnet.EventHandler.
func (n *tcpCore) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	log.Println("client: OnClose")
	return
}

// OnOpen implements gnet.EventHandler.
func (n *tcpCore) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Println("client: OnOpen")
	return
}

// OnShutdown implements gnet.EventHandler.
func (n *tcpCore) OnShutdown(eng gnet.Engine) {
	log.Println("client: OnShutdown")
}

// OnTick implements gnet.EventHandler.
func (n *tcpCore) OnTick() (delay time.Duration, action gnet.Action) {
	return
}

// OnTraffic implements gnet.EventHandler.
func (n *tcpCore) OnTraffic(c gnet.Conn) (action gnet.Action) {
	var buf bytes.Buffer
	m, err := c.WriteTo(&buf)
	if err != nil {
		return
	}
	if m > 0 {
		n.recv <- buf.Bytes()
	}
	return
}
