//go:build cgo && (dragonfly || freebsd || netbsd || openbsd)

package systemops

import "net"

func init() {
	testCases = append(testCases, testCase{
		name:              "To more specific route without custom dialer via vpn",
		expectedInterface: expectedVPNint,
		dialer:            &net.Dialer{},
		expectedPacket:    createPacketExpectation("100.64.0.1", 12345, "10.10.0.2", 53),
	})
}
