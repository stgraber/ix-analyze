package main

import (
	"encoding/csv"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type peer struct {
	ASN             string
	Label           string
	Company         string
	HWAddr          string
	IPv4            string
	IPv6            string
	RouteServerIPv4 bool
	RouteServerIPv6 bool
	LinkSpeed       string
	IRR             string
	Note            string
}

var members map[string]*peer
var overrides map[string]*peer

func loadMembers(fileName string) error {
	return loadFile(fileName, &members)
}

func loadOverrides(fileName string) error {
	return loadFile(fileName, &overrides)
}

func loadFile(fileName string, target *map[string]*peer) error {
	// Parse the current file.
	entries, err := parseFile(fileName)
	if err != nil {
		return err
	}

	// Update the global data.
	*target = entries

	// Setup watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			event, ok := <-watcher.Events
			if !ok {
				return
			}

			if !event.Has(fsnotify.Write) || event.Name != fileName {
				continue
			}

			// Refresh the list.
			entries, err := parseFile(fileName)
			if err != nil {
				continue
			}

			*target = entries
		}
	}()

	err = watcher.Add(filepath.Dir(fileName))
	if err != nil {
		return err
	}

	return nil
}

func parseFile(fileName string) (map[string]*peer, error) {
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

	entries := map[string]*peer{}
	for _, record := range records {
		entry := &peer{
			ASN:             record[0],
			Label:           record[1],
			Company:         record[2],
			HWAddr:          record[3],
			IPv4:            record[4],
			IPv6:            record[5],
			RouteServerIPv4: record[6] == "Yes",
			RouteServerIPv6: record[7] == "Yes",
			LinkSpeed:       record[8],
			IRR:             record[9],
		}

		if len(record) > 10 {
			entry.Note = record[10]
		}

		entries[entry.HWAddr] = entry
	}

	return entries, nil
}
