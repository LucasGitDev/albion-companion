//go:build linux

// Copied from fork of ao-data/albiondata-client (MIT, see NOTICE).

package capture

import (
	"net"
)

// Gets all physical interfaces based on filter results, ignoring all VM, Loopback and Tunnel interfaces.
func getAllPhysicalInterface() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var outInterfaces []string

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 && isPhysicalInterface(iface.HardwareAddr.String()) && ifaceHasUsableIP(iface) {
			outInterfaces = append(outInterfaces, iface.Name)
		}
	}

	return outInterfaces, nil
}
