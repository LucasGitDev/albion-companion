//go:build darwin

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

	for _, _interface := range interfaces {
		if _interface.Flags&net.FlagLoopback == 0 && _interface.Flags&net.FlagUp != 0 && isPhysicalInterface(_interface.HardwareAddr.String()) && ifaceHasUsableIP(_interface) {
			outInterfaces = append(outInterfaces, _interface.Name)
		}
	}

	return outInterfaces, nil
}
