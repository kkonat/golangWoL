package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"sort"
	"strconv"

	"github.com/kkonat/WoL/Internal/wol"
	//"../wol"
)

func main() {
	const usage string = `usage:
	wol list
		 lists all available interfaces
	wol wake interface_no macAddress
		wakes specific device using the interface with the given interface_no`

	// Obtain a list of network interfaces
	ifs, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	// get program arguments slice
	args := os.Args[1:]

	// no arguments, print usage
	if len(args) == 0 {
		fmt.Println(usage)
		return
	}

	// process args
	switch args[0] {
	case "list":
		// sort the list for nicer output
		sort.Slice(ifs, func(i, j int) bool { return ifs[i].Index < ifs[j].Index })

		// print interfaces table
		fmt.Println("interface_no\t IP address\t Interface name")
		fmt.Println("============\t ============= \t =============================")

		for _, interf := range ifs {
			localAddr, err := wol.IfToIp(interf.Name)
			if err == nil {
				fmt.Println(interf.Index, "\t\t", localAddr.IP, "\t", interf.Name)
			}
		}
		return
	case "wake":
		// wake the selected host

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
					wol.Wake(interf.Name, args[2])
					found = true
				}
			}
			if !found {
				fmt.Println("interface_no not found, use wol list to list valid numbers")
			}

		} else {
			fmt.Println("you must provide interface_no and mac address")
		}
	default:
		fmt.Println(usage)
	}
}
