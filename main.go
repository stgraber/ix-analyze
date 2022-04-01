package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/lxc/lxd/shared/units"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type counter struct {
	name string
	rx   int64
	tx   int64
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
		table.SetHeader([]string{"PEER", "RX", "TX", "TOTAL"})
		table.SetBorder(false)
		table.SetAutoWrapText(false)

		data := traffic.toSlice()
		sort.SliceStable(data, func(i int, j int) bool {
			return data[i].total > data[j].total
		})

		for _, entry := range data {
			table.Append([]string{
				entry.name,
				units.GetByteSizeString(entry.rx, 2),
				units.GetByteSizeString(entry.tx, 2),
				units.GetByteSizeString(entry.total, 2),
			})
		}

		table.Render()
	}
}

func run() error {
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
		if layer != nil {
			ether := layer.(*layers.Ethernet)

			src := ether.SrcMAC.String()
			if members[src] != "" {
				src = members[src]
			}
			if traffic[src] == nil {
				traffic[src] = &counter{name: src}
			}

			dst := ether.DstMAC.String()
			if members[dst] != "" {
				src = members[dst]
			}
			if traffic[dst] == nil {
				traffic[dst] = &counter{name: dst}
			}

			traffic[src].rx += length
			traffic[src].total += length
			traffic[dst].tx += length
			traffic[dst].total += length
		}
	}
}
