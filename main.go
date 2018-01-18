package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/mushroomsir/container/network"
)

func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	case "bridge":
		createBridge(os.Args[2])
	default:
		panic("wat should I do")
	}
}

func createBridge(pid string) {

	bridgeIP, bridgeSubnet, err := net.ParseCIDR("10.10.10.1/24")
	check(err)

	containerIP, _, err := net.ParseCIDR("10.10.10.2/24")
	check(err)

	netConfig := network.NetworkConfig{
		BridgeName:     "brgtest",
		BridgeIP:       bridgeIP,
		ContainerIP:    containerIP,
		Subnet:         bridgeSubnet,
		VethNamePrefix: "vethtest",
	}
	processID, err := strconv.Atoi(pid)
	check(err)

	bridge := network.NewBridge()
	veth := network.NewVeth()
	network.ApplyHost(brdige, veth, netConfig, processID)
	network.ApplyContainer(netConfig, processID)
}
func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	log.Println(cmd.Process.Pid)
}

func child() {
	syscall.Sethostname([]byte("container" + strconv.Itoa(os.Getgid())))
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	log.Println("app:", cmd.Process.Pid)
}

func check(err error) {
	if err != nil {
		fmt.Printf("ERROR - %s\n", err.Error())
		os.Exit(1)
	}
}
