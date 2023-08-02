package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func isPrivateMAC(mac string) bool {
	// Broadcast.
	if mac == "ff:ff:ff:ff:ff:ff" {
		return true
	}

	// IPv6 multicast.
	if strings.HasPrefix(mac, "33:33:") {
		return true
	}

	return false
}

func run() error {
	// Usage.
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <interface> <members CSV> [override CSV]\n", os.Args[0])
		return nil
	}

	// Setup packet capture.
	handle, err := pcap.OpenLive(os.Args[1], 65536, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	// Load member list.
	err = loadMembers(os.Args[2])
	if err != nil {
		return err
	}

	if len(os.Args) >= 4 {
		err = loadOverrides(os.Args[3])
		if err != nil {
			return err
		}
	}

	// Prepare counters.
	traffic := trafficCounter{}

	// Table rendering.
	go renderTable(traffic)

	for {
		data, _, err := handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return err
		}

		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		length := int64(len(packet.Data()))
		layer := packet.Layer(layers.LayerTypeEthernet)

		isIPv6 := packet.Layer(layers.LayerTypeIPv6)
		if layer != nil {
			ether := layer.(*layers.Ethernet)

			src := ether.SrcMAC.String()
			if isPrivateMAC(src) {
				continue
			}

			if traffic[src] == nil {
				traffic[src] = &counter{hwaddr: src}
			}

			dst := ether.DstMAC.String()
			if isPrivateMAC(dst) {
				continue
			}

			if traffic[dst] == nil {
				traffic[dst] = &counter{hwaddr: dst}
			}

			traffic[src].v4rx += length
			if isIPv6 != nil {
				traffic[src].v6rx += length
			} else {
				traffic[src].v4rx += length
			}
			traffic[src].total += length

			if isIPv6 != nil {
				traffic[dst].v6tx += length
			} else {
				traffic[dst].v4tx += length
			}
			traffic[dst].total += length
		}
	}
}
