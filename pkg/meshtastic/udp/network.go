package udp

import (
	"errors"
	"fmt"
	"net"
)

// resolveLocalAddress determines the local UDP address and the corresponding network interface
// that would be used to communicate with a remote host and port.
func resolveLocalAddress(remoteHost string, remotePort int) (*net.UDPAddr, net.Interface, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(remoteHost, remotePort))
	if err != nil {
		return nil, net.Interface{}, err
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, net.Interface{}, err
	}

	laddr := conn.LocalAddr().(*net.UDPAddr)

	if err = conn.Close(); err != nil {
		return nil, net.Interface{}, err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, net.Interface{}, err
	}

	for _, iface := range ifaces {
		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				if addrNet, ok := addr.(*net.IPNet); ok && laddr.IP.Equal(addrNet.IP) {
					return laddr, iface, nil
				}
			}
		}
	}

	return laddr, net.Interface{}, errors.New("could not find matching interface for address")
}
