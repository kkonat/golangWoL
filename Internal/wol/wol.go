package wol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

type macAddress [6]byte
type magicPacket struct {
	header  [6]byte
	payload [16]macAddress
}

// IftoIp Gets returns UDP IP address of an interface
func IfToIp(iface string) (*net.UDPAddr, error) {

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

func buildPacket(mac string) (*magicPacket, error) {
	var packet magicPacket
	var macAddr macAddress

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

func Wake(bcastInterface string, macAddr string) {

	const bcastAddr string = "255.255.255.255:9"

	// get local IP
	localAddr, err := IfToIp(bcastInterface)
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
