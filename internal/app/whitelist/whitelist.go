package whitelist

import (
	"fmt"
	"net"
	"net/http"
)

// WhiteList struct with white list data
type WhiteList struct {
	subnet *net.IPNet
}

// NewWhiteList creates new WhiteList
func NewWhiteList(subnet *net.IPNet) *WhiteList {
	return &WhiteList{subnet: subnet}
}

// Handler WhiteList middlewares handler
func (wl *WhiteList) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := net.ParseIP(r.Header.Get("X-Real-IP"))
		if wl.subnet == nil || !wl.subnet.Contains(ip) {
			http.Error(w, fmt.Sprintf("ip %s is not from trusted subnet", ip.String()), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
