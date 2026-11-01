// Copied from fork of ao-data/albiondata-client (MIT, see NOTICE).
package capture

import (
	"net"
	"strings"
)

// ifaceHasUsableIP reports whether iface has at least one assigned IP that is a
// plausible capture target: not loopback, link-local, or unspecified. macOS
// creates many address-less placeholder interfaces (anpi*, spare enN, idle
// bridge*/vmenet*, down utun*); opening a pcap handle for each one consumes a
// /dev/bpf* device, and since grant-capture only chowns the devices that
// existed when it ran, the real interface can land on a later, root-owned bpf
// and fail with "permission to capture on that device" (TASK-38). Requiring an
// assigned IP keeps the set to the interfaces that can actually carry traffic.
func ifaceHasUsableIP(iface net.Interface) bool {
	addrs, err := iface.Addrs()
	if err != nil {
		return false
	}
	for _, a := range addrs {
		var ip net.IP
		switch v := a.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if isUsableCaptureIP(ip) {
			return true
		}
	}
	return false
}

// isUsableCaptureIP reports whether an assigned address marks an interface worth
// opening: a real, routable-looking IP rather than loopback, link-local, or the
// unspecified address.
func isUsableCaptureIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return false
	}
	return true
}

// Mac Address parts to look for, and identify non physical devices. There may be more, update me!
var macAddrPartsToFilter = []string{
	"ac:de:48:00:11:22", // Mac OS touchbar (en5)
	"00:03:FF",          // Microsoft Hyper-V, Virtual Server, Virtual PC
	"00:15:5D",          // Windows WSL2 or Hyper-V related
	"0A:00:27",          // VirtualBox
	"00:00:00:00:00",    // Teredo Tunneling Pseudo-Interface
	"00:50:56",          // VMware ESX 3, Server, Workstation, Player
	"00:1C:14",          // VMware ESX 3, Server, Workstation, Player
	"00:0C:29",          // VMware ESX 3, Server, Workstation, Player
	"00:05:69",          // VMware ESX 3, Server, Workstation, Player
	"00:1C:42",          // Microsoft Hyper-V, Virtual Server, Virtual PC
	"00:0F:4B",          // Virtual Iron 4
	"00:16:3E",          // Red Hat Xen, Oracle VM, XenSource, Novell Xen
	"08:00:27",          // Sun xVM VirtualBox
}

// Filters the possible physical interface address by comparing it to known popular VM Software adresses
// and Teredo Tunneling Pseudo-Interface.
func isPhysicalInterface(addr string) bool {
	for _, macPart := range macAddrPartsToFilter {
		if strings.HasPrefix(strings.ToLower(addr), strings.ToLower(macPart)) {
			return false
		}
	}

	return true
}
