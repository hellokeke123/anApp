package model

import "net"

const CLIENT = "client"

const UDP = "udp"
const TCP = "tcp"
const DNS = "dns"

const DNS_PORT = 53

const PROC_DAEMON = "daemon"

var DEFAULT_IPNET = net.IPNet{IP: net.IPv4zero, Mask: net.IPv4Mask(0, 0, 0, 0)}
