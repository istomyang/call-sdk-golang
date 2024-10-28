package net_core

type NetCore interface {
	Send(data []byte)
	Recv() <-chan []byte
	Run() error
}
