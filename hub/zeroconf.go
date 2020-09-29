package hub

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
)

func ZeroConfServer(ctx context.Context, name string, address string) {
	logrus.Infof("Starting MDNS server")
	ifaces, err := net.Interfaces()
	if err != nil {
		logrus.Errorf("getting interfaces failed: %v", err)
	}

	allowedInterfaces := []net.Interface{}
	for _, interf := range ifaces {
		addresses, _ := interf.Addrs()
		for _, addr := range addresses {
			if strings.Contains(addr.String(), address) {
				allowedInterfaces = append(allowedInterfaces, interf)
				break
			}
		}
	}

	serviceName := "_automationhub._tcp"
	if name != "" {
		serviceName = fmt.Sprintf("_automationhub_%s._tcp", name)
	}

	server, err := zeroconf.Register("AutomationHub", serviceName, "local.", 3939, []string{"txtv=0", "lo=1", "la=2"}, allowedInterfaces)
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	// Clean exit.
	select {
	case <-ctx.Done():
	}

	logrus.Infof("MDNS server shutdown.")
}
