// Package middleware provides various middleware functionality.
package middleware

import (
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/danilovkiri/dk_go_url_shortener/internal/config"
)

// TrustedNetHandler sets object structure.
type TrustedNetHandler struct {
	Resolved bool
	IP       net.IP
	IPNet    *net.IPNet
}

// NewTrustedNetHandler initializes a new trusted network handler.
func NewTrustedNetHandler(cfg *config.Config) *TrustedNetHandler {
	ip, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	log.Print(err)
	var parseError *net.ParseError
	if errors.As(err, &parseError) {
		log.Println("Trusted network was not initialized:", err)
		return &TrustedNetHandler{
			Resolved: false,
			IP:       nil,
			IPNet:    nil,
		}
	}
	return &TrustedNetHandler{
		Resolved: true,
		IP:       ip,
		IPNet:    ipnet,
	}
}

// TrustedNetworkHandler provides trusted network handling functionality.
func (tn *TrustedNetHandler) TrustedNetworkHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tn.Resolved {
			http.Error(w, "Internal subnet access violation", http.StatusForbidden)
			return
		}
		addr := r.RemoteAddr
		ipStr, _, err := net.SplitHostPort(addr)
		ipMain := net.ParseIP(ipStr)
		if err != nil || ipMain == nil || !tn.IPNet.Contains(ipMain) {
			ipStrFwd := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStrFwd)
			if ip == nil {
				ips := r.Header.Get("X-Forwarded-For")
				ipStrs := strings.Split(ips, ",")
				ipStr = ipStrs[0]
				ip = net.ParseIP(ipStr)
			}
			if ip == nil || !tn.IPNet.Contains(ip) {
				http.Error(w, "Internal subnet access violation", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
