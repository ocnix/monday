package proxy

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

func getInterface() (string, error) {
	var networkInterfaceName string
	ifaces, err := net.Interfaces()
	if err != nil{
		fmt.Println("Error getting interfaces: ", err)
		return "", err
	}
	for _, i := range ifaces{
		ifaceFlags := i.Flags.String()
		if strings.Contains(ifaceFlags, "loopback"){
			networkInterfaceName = i.Name
			return networkInterfaceName, nil
		} else {
			return "", err
		}

	}

	return "", nil
}


func generateIP(a byte, b byte, c byte, d int, port string) (net.IP, error) {
	ip := net.IPv4(a, b, c, byte(d))

	networkInterface, err := getInterface()
	if err != nil{
		return net.IP{}, err
	}

	fmt.Println("Found interface name: ", networkInterface)
	iface, err := net.InterfaceByName(networkInterface)

	if err != nil {
		return net.IP{}, err
	}

	//Interface information
	/*
	fmt.Println("Interface index: ", iface.Index)
	fmt.Println("Interface name: ", iface.Name)
	fmt.Println("Interface flags", iface.Flags)
	fmt.Println("Interface MTU", iface.MTU)
	fmt.Println("Interface Hardware address", iface.HardwareAddr)
	*/

	for i := d; i < 255; i++ {
		ip = net.IPv4(a, b, c, byte(i))
		fmt.Println("IP address created: ", ip)

		addrs, err := iface.Addrs()
		if err != nil {
			return net.IP{}, err
		}
		fmt.Println("Interface ip address: ", addrs[0].String())

		// Try to assign port to the IP addresses already assigned to the interface
		for _, addr := range addrs {
			fmt.Printf("generated ip: %v/8 ----- interface ip: %v\n", ip.String(), addr.String())
			if addr.String() == ip.String()+"/8" {
				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ip.String(), port))
				if err != nil {
					return net.IPv4(a, b, c, byte(i)), nil
				}
				conn.Close()
			} else{
				os.Exit(1)
			}
		}

		// Add a new IP address on the network interface
		command := "ifconfig"
		args := []string{iface.Name, ip.String(), "up"}
		if err := exec.Command(command, args...).Run(); err != nil {
			return net.IP{}, fmt.Errorf("Cannot run ifconfig command to add new IP address (%s) on lo0 interface: %v", ip.String(), err)
		}

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ip.String(), port))
		if err == nil {
			return net.IPv4(a, b, c, byte(i)), nil
		}
		if conn != nil {
			conn.Close()
		}
	}

	return net.IP{}, fmt.Errorf("Unable to find an available IP/Port (ip: %d.%d.%d.%d:%s)", a, b, c, d, port)
}
