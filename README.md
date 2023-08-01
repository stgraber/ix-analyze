# ix-analyze

Small command line tool which looks at traffic on a peering interface
and lines up the IX peers based on their MAC addresses. This is useful
to see what peer you interact with the most, get traffic totals for IPv4
and IPv6 traffic and identify where direct peering may make sense.

This tool requires that the router be a Linux system, it doesn't support
fetching network traffic data from a separate router.

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
