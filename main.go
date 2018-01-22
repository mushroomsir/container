package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"

	"github.com/mushroomsir/container/network"
)

var netConfig network.NetworkConfig

func init() {
	bridgeIP, bridgeSubnet, err := net.ParseCIDR("10.10.10.1/24")
	check(err)

	containerIP, _, err := net.ParseCIDR("10.10.10.2/24")
	check(err)

	netConfig = network.NetworkConfig{
		BridgeName:     "brg0",
		BridgeIP:       bridgeIP,
		ContainerIP:    containerIP,
		Subnet:         bridgeSubnet,
		VethNamePrefix: "veth",
	}

}
func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("wat should I do")
	}
}

func createBridge(pid int) {

	bridge := network.NewBridge()
	veth := network.NewVeth()
	err := network.ApplyHost(bridge, veth, netConfig, pid)
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

	iptablesRules := network.GetIptablesRules("10.10.10.1/24", "eth0", "container0")
	if err := network.SetIptables(iptablesRules); err != nil {
		log.Println(fmt.Errorf("set iptables err: %v", err))
	}

	createBridge(cmd.Process.Pid)

	err = cmd.Wait()
	check(err)
}

func child() {
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	initContainer()

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	log.Println("app:", cmd.Process.Pid)
}

func initContainer() {
	name := "container" + strconv.Itoa(os.Getgid())
	syscall.Sethostname([]byte(name))

	pwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("get pwd err: %v\n", err))
	}
	target := path.Join(pwd, "rootfs")
	if err := syscall.Chroot(target); err != nil {
		panic(fmt.Sprintf("chroot err: %v\n", err))
	}
	if err := os.Chdir("/"); err != nil {
		panic(fmt.Sprintf("chdir err: %v\n", err))
	}

	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		panic(fmt.Sprintf("failed to mount proc to %s: %v", target, err))
	}

	err = network.ApplyContainer(netConfig)
	check(err)
}

func check(err error) {
	if err != nil {
		//fmt.Printf("ERROR - %s\n", err.Error())
		panic(err)
	}
}
