package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/canonical/lxd/shared/units"
	"github.com/olekukonko/tablewriter"
)

func renderTable(traffic trafficCounter) {
	for {
		time.Sleep(3 * time.Second)
		fmt.Print("\033[H\033[2J")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"PEER", "IPV4 RX", "IPV4 TX", "IPV6 RX", "IPV6 TX", "TOTAL", "NOTE"})
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

			var name string
			if overrides[entry.hwaddr] != nil && overrides[entry.hwaddr].Company != "" {
				name = overrides[entry.hwaddr].Company
			} else if members[entry.hwaddr] != nil {
				name = members[entry.hwaddr].Company
			} else {
				name = fmt.Sprintf("UNKNOWN (%s)", entry.hwaddr)
			}

			var asn string
			if overrides[entry.hwaddr] != nil && overrides[entry.hwaddr].ASN != "" {
				asn = overrides[entry.hwaddr].ASN
			} else if members[entry.hwaddr] != nil {
				asn = members[entry.hwaddr].ASN
			}

			if asn != "" {
				name = fmt.Sprintf("%s (%s)", name, asn)
			}

			var note string
			if overrides[entry.hwaddr] != nil && overrides[entry.hwaddr].Note != "" {
				note = overrides[entry.hwaddr].Note
			}

			table.Append([]string{
				name,
				units.GetByteSizeString(entry.v4rx, 2),
				units.GetByteSizeString(entry.v4tx, 2),
				units.GetByteSizeString(entry.v6rx, 2),
				units.GetByteSizeString(entry.v6tx, 2),
				units.GetByteSizeString(entry.total, 2),
				note,
			})
		}

		table.Render()
	}
}
