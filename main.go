package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/canonical/lxd/shared/units"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/olekukonko/tablewriter"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type counter struct {
	name  string
	v4rx  int64
	v4tx  int64
	v6rx  int64
	v6tx  int64
	total int64
}

type trafficCounter map[string]*counter

func (tc trafficCounter) toSlice() []*counter {
	out := make([]*counter, 0, len(tc))
	for _, v := range tc {
		out = append(out, v)
	}

	return out
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

func getMembers(fileName string) (map[string]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	members := map[string]string{}
	for _, record := range records {
		members[record[3]] = fmt.Sprintf("%s (%s)", record[2], record[0])
	}

	return members, nil
}

func renderTable(traffic trafficCounter) {
	for {
		time.Sleep(3 * time.Second)
		fmt.Print("\033[H\033[2J")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"PEER", "IPV4 RX", "IPV4 TX", "IPV6 RX", "IPV6 TX", "TOTAL"})
		table.SetBorder(false)
		table.SetAutoWrapText(false)

		data := traffic.toSlice()
		sort.SliceStable(data, func(i int, j int) bool {
			return data[i].total > data[j].total
		})

		for _, entry := range data {
			// Ignore anything that's got less than 1MiB of traffic.
			if entry.total < 1024*1024 {
				continue
			}

			table.Append([]string{
				entry.name,
				units.GetByteSizeString(entry.v4rx, 2),
				units.GetByteSizeString(entry.v4tx, 2),
				units.GetByteSizeString(entry.v6rx, 2),
				units.GetByteSizeString(entry.v6tx, 2),
				units.GetByteSizeString(entry.total, 2),
			})
		}

		table.Render()
	}
}

func run() error {
	// Usage.
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <interface> <CSV file>\n", os.Args[0])
		return nil
	}

	// Setup packet capture.
	handle, err := pcap.OpenLive(os.Args[1], 65536, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	// Load member list.
	members, err := getMembers(os.Args[2])
	if err != nil {
		return err
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
			if members[src] != "" {
				src = members[src]
			} else if isPrivateMAC(src) {
				continue
			} else {
				src = fmt.Sprintf("UNKNOWN (%s)", src)
			}
			if traffic[src] == nil {
				traffic[src] = &counter{name: src}
			}

			dst := ether.DstMAC.String()
			if members[dst] != "" {
				dst = members[dst]
			} else if isPrivateMAC(dst) {
				continue
			} else {
				dst = fmt.Sprintf("UNKNOWN (%s)", dst)
			}
			if traffic[dst] == nil {
				traffic[dst] = &counter{name: dst}
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
