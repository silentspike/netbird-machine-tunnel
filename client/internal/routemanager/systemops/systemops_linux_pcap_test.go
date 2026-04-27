//go:build cgo && !android

package systemops

import "net"

func init() {
	testCases = append(testCases, testCase{
		name:              "To more specific route without custom dialer via physical interface",
		expectedInterface: expectedInternalInt,
		dialer:            &net.Dialer{},
		expectedPacket:    createPacketExpectation("192.168.1.1", 12345, "10.10.0.2", 53),
	})
}
