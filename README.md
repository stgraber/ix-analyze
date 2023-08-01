# ix-analyze

Small command line tool which looks at traffic on a peering interface
and lines up the IX peers based on their MAC addresses. This is useful
to see what peer you interact with the most, get traffic totals for IPv4
and IPv6 traffic and identify where direct peering may make sense.

This tool requires that the router be a Linux system, it doesn't support
fetching network traffic data from a separate router.

## Output
Here is some example output as seen after a few minutes of capture on the [Zabbly](https://zabbly.com) router.

```
               PEER               | IPV4 RX  | IPV4 TX  | IPV6 RX  | IPV6 TX  |  TOTAL
----------------------------------+----------+----------+----------+----------+----------
  Zabbly (399760)                 | 3.69GB   | 23.21MB  | 2.14GB   | 738.98MB | 3.68GB
  Hurricane Electric (6939)       | 92.84MB  | 181.09MB | 68.42MB  | 2.13GB   | 2.39GB
  Google (B) (15169)              | 681.03MB | 579.81MB | 669.60MB | 4.26MB   | 1.26GB
  OVH Infrastructures inc (16276) | 2.37MB   | 4.44MB   | 433.97kB | 4.21MB   | 10.05MB
  CloudFlare (13335)              | 315.37kB | 1.39MB   | 145.24kB | 3.09MB   | 4.72MB
  METROOPTIC (29909)              | 2.06kB   | 3.48MB   | 500B     | 328B     | 3.48MB
  Google (15169)                  | 5.34MB   | 38.94kB  | 328B     | 328B     | 2.71MB
  TekSavvy Solutions Inc (5645)   | 467.62kB | 1.82MB   | 10.17kB  | 18.87kB  | 2.07MB
```

## Installing (Go)
go install -v github.com/stgraber/ix-analyze@latest

## Installing (snap)
A snap package is available: `snap install ix-analyze`

## Input file
The input file is expected to be a CSV file with the following columns:

 - ASN
 - Member name (short) (unused)
 - Member name (long)
 - MAC address
 - IPv4 address (unused)
 - IPv6 address (unused)
 - Member of IPv4 route server (unused)
 - Member of IPv6 route server (unused)
 - Link speed (unused)
 - IRR record (unused)

The tool was developed for use with the [Montreal Internal Exchange](https://qix.ca) and therefore uses its format.
