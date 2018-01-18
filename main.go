package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/mushroomsir/container/netns"
	"github.com/mushroomsir/container/network"
)

func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	// case "bridge":
	// 	createBridge(os.Args[2])
	default:
		panic("wat should I do")
	}
}

func createBridge(pid int) {
	//log.Println(pid)
	bridgeIP, bridgeSubnet, err := net.ParseCIDR("10.10.10.1/24")
	check(err)

	containerIP, _, err := net.ParseCIDR("10.10.10.2/24")
	check(err)

	netConfig := network.NetworkConfig{
		BridgeName:     "brg0",
		BridgeIP:       bridgeIP,
		ContainerIP:    containerIP,
		Subnet:         bridgeSubnet,
		VethNamePrefix: "veth",
	}
	// processID, err := strconv.Atoi(pid)
	// check(err)

	bridge := network.NewBridge()
	veth := network.NewVeth()
	err = network.ApplyHost(bridge, veth, netConfig, pid)
	check(err)

	netnsExecer := &netns.Execer{}

	err = network.ApplyContainer(netConfig, pid, netnsExecer)
	check(err)
}
func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	check(err)

	log.Println(cmd.Process.Pid)
	createBridge(cmd.Process.Pid)

	err = cmd.Wait()
	check(err)
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
		//fmt.Printf("ERROR - %s\n", err.Error())
		panic(err)
	}
}
