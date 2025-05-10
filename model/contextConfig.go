package model

import "fmt"

type ContextConfig struct {
	Title         string
	App           App
	InitNet       InitNet
	ContextServer ContextServer
	ContextClient ContextClient
}

type App struct {
	Context string
}

type InitNet struct {
	Ip string
}

type ContextServer struct {
	Port string
	Ip   string
	Path string
	Key  string
}

type ContextClient struct {
	Direct                    bool
	EnableEnforceDns          bool
	EnforceDOH                string
	ReadEnableEnforceDnsChan  chan bool
	WriteEnableEnforceDnsChan chan bool
	SourceIp                  string
	TargetIp                  string
	TunIp                     string
}

func (client ContextClient) IsEnableEnforceDns(port uint16) bool {
	return DNS_PORT == port && client.EnableEnforceDns && len(ContextConfigImp.ContextClient.EnforceDOH) > 0
}

func (client ContextClient) GetEnableEnforceDHO() string {
	return fmt.Sprint(ContextConfigImp.ContextClient.EnforceDOH)
}
