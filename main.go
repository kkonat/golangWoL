package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func ipFromInterface(iface string) (*net.UDPAddr, error) {

	ief, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, err
	}

	addrs, err := ief.Addrs()
	if err == nil && len(addrs) <= 0 {
		err = fmt.Errorf("no address associated with interface %s", iface)
	}
	if err != nil {
		return nil, err
	}

	// Validate that one of the addrs is a valid network IP address.
	for _, addr := range addrs {
		switch ip := addr.(type) {
		case *net.IPNet:
			if !ip.IP.IsLoopback() && ip.IP.To4() != nil {
				return &net.UDPAddr{
					IP: ip.IP,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("no address associated with interface %s", iface)
}

type MACAddress [6]byte
type MagicPacket struct {
	header  [6]byte
	payload [16]MACAddress
}

func buildPacket(mac string) (*MagicPacket, error) {
	var packet MagicPacket
	var macAddr MACAddress

	hwAddr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	// Copy bytes from the returned HardwareAddr -> a fixed size MACAddress.
	for idx := range macAddr {
		macAddr[idx] = hwAddr[idx]
	}

	// header  6 x 0xFF.
	for idx := range packet.header {
		packet.header[idx] = 0xFF
	}

	// payload:  16 x MAC addr.
	for idx := range packet.payload {
		packet.payload[idx] = macAddr
	}

	return &packet, nil
}

func wake(bcastInterface string, macAddr string) {
	//bcastInterface := "Ethernet 2"
	//	macAddr := "14:DA:E9:03:FD:AC"
	bcastAddr := "255.255.255.255:9"

	localAddr, err := ipFromInterface(bcastInterface)
	if err != nil {
		log.Fatal(err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", bcastAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Build the magic packet.
	mp, err := buildPacket(macAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare  bytes to send

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, mp); err != nil {
		log.Fatal(err)
	}

	// setup  UDP connection
	conn, err := net.DialUDP("udp", localAddr, udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Printf("Waking MAC %s\n", macAddr)
	fmt.Printf("... Broadcasting to: %s\n", bcastAddr)
	n, err := conn.Write(buf.Bytes())
	if err == nil && n != 102 {
		err = fmt.Errorf("magic packet sent was %d bytes (expected 102 bytes sent)", n)
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Magic packet sent successfully to %s\n", macAddr)
}
func main() {
	const usage string = `usage:
	wol list
		 lists all available interfaces
	wol wake interface_no macAddress
		wakes specific device using the interface with the given interface_no`

	ifs, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(usage)
	} else if args[0] == "list" {

		fmt.Println("interface_no\tName")
		for _, interf := range ifs {
			fmt.Println(interf.Index, "\t\t", interf.Name)
		}
		return
	} else if args[0] == "wake" {
		if len(args) == 3 {
			index, err := strconv.Atoi(args[1])
			if err != nil {
				fmt.Println("wrong index format")
				return
			}
			found := false
			for _, interf := range ifs {
				if interf.Index == index {
					fmt.Println("Waking using interface no. ", args[1])
					wake(interf.Name, args[2])
					found = true
					break
				}
			}
			if !found {
				fmt.Println("interface_no not found, use wol list to list valid numbers")
				return
			}

		} else {
			fmt.Println("you must provide interface_no and mac address")
			return
		}
	} else {
		fmt.Println(usage)
		return
	}

}
