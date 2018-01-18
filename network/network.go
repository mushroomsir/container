package network

import (
	"fmt"
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

type NetworkConfig struct {
	BridgeName     string
	BridgeIP       net.IP
	ContainerIP    net.IP
	Subnet         *net.IPNet
	VethNamePrefix string
}

func ApplyHost(bridge *Bridge, veth *Veth, netConfig NetworkConfig, pid int) error {
	b, err := bridge.Create(netConfig.BridgeName, netConfig.BridgeIP, netConfig.Subnet)
	if err != nil {
		return err
	}

	hostVeth, containerVeth, err := veth.Create(netConfig.VethNamePrefix)
	if err != nil {
		return err
	}

	err = bridge.Attach(b, hostVeth)
	if err != nil {
		return err
	}

	err = veth.MoveToNetworkNamespace(containerVeth, pid)
	if err != nil {
		return err
	}

	return nil
}
func ApplyContainer(netConfig NetworkConfig) error {

	time.Sleep(200 * time.Millisecond)

	//cbFunc := func() error {
	containerVethName := fmt.Sprintf("%s1", netConfig.VethNamePrefix)
	link, err := netlink.LinkByName(containerVethName)
	if err != nil {
		return fmt.Errorf("Container veth '%s' not found", containerVethName)
	}

	addr := &netlink.Addr{IPNet: &net.IPNet{IP: netConfig.ContainerIP, Mask: netConfig.Subnet.Mask}}
	err = netlink.AddrAdd(link, addr)
	if err != nil {
		return fmt.Errorf("Unable to assign IP address '%s' to %s", netConfig.ContainerIP, containerVethName)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	route := &netlink.Route{
		Scope:     netlink.SCOPE_UNIVERSE,
		LinkIndex: link.Attrs().Index,
		Gw:        netConfig.BridgeIP,
	}

	return netlink.RouteAdd(route)
}
