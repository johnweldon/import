package pkg

import (
	"log"
	"net"
	"net/http"
	"strings"
)

func GetNetworks(cidrs []string) []*net.IPNet {
	var safe []*net.IPNet
	hash := map[string]*net.IPNet{}
	for _, s := range cidrs {
		ip := cleanIP(s)
		_, n, err := net.ParseCIDR(ip)
		if err != nil {
			log.Printf("ERROR: invalid safe ip %q: %v", ip, err)
			continue
		}
		if _, ok := hash[ip]; !ok {
			hash[ip] = n
			safe = append(safe, n)
		}
	}
	return safe
}

func queryIP(r *http.Request) string {
	vars := r.URL.Query()
	ip := vars.Get("ip")
	if ip == "" {
		ip = remoteIP(r)
	}
	return ip
}

func lastIP(r *http.Request) string {
	return cleanIP(r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")])
}

func lastForwarder(r *http.Request) string {
	var ip string
	if f := forwarders(r); len(f) > 0 {
		// get LAST forwarder
		ip = f[len(f)-1]
	}
	if ip == "" {
		ip = lastIP(r)
	}
	return cleanIP(ip)
}

func remoteIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		if f := forwarders(r); len(f) > 0 {
			// get FIRST forwarder
			ip = f[0]
		}
	}
	if ip == "" {
		ip = lastIP(r)
	}
	return cleanIP(ip)
}

func forwarders(r *http.Request) []string {
	var f []string
	if forwards, ok := r.Header["X-Forwarded-For"]; ok {
		for _, fw := range forwards {
			for _, s := range strings.Split(fw, ",") {
				if ip := cleanIP(s); ip != "" {
					f = append(f, ip)
				}
			}
		}
	}
	return f
}

func cleanIP(ip string) string {
	return strings.Map(
		func(r rune) rune {
			switch r {
			case
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
				'A', 'B', 'C', 'D', 'E', 'F',
				'a', 'b', 'c', 'd', 'e', 'f',
				':', '.', '/':
				return r
			default:
				return -1
			}
		}, ip)
}
