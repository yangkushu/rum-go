package utils

import (
	"github.com/yangkushu/rum-go/log"
	"net"
)

var (
	localIP     = ""
	privateCIDR []*net.IPNet
)

// getFaces return addresses from interfaces that is up
func getFaces() ([]net.Addr, error) {
	var upAddrs []net.Addr

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Error("get net Interfaces failed", log.String("error", err.Error()))
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if (iface.Flags & net.FlagLoopback) != 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			log.Error("get InterfaceAddress failed", log.String("error", err.Error()))
			return nil, err
		}

		upAddrs = append(upAddrs, addresses...)
	}

	return upAddrs, nil
}

func isFilteredIP(ip net.IP) bool {
	for _, privateIP := range privateCIDR {
		if privateIP.Contains(ip) {
			return true
		}
	}
	return false
}

func GetLocalIp() string {
	if localIP != "" {
		return localIP
	}

	faces, err := getFaces()
	if err != nil {
		return ""
	}

	for _, address := range faces {
		ipNet, ok := address.(*net.IPNet)
		if !ok || ipNet.IP.To4() == nil || isFilteredIP(ipNet.IP) {
			continue
		}

		localIP = ipNet.IP.String()
		break
	}

	return localIP
}
