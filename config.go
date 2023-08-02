package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

var members map[string]string

func loadMembers(fileName string) error {
	// Parse the current file.
	entries, err := parseMembers(fileName)
	if err != nil {
		return err
	}

	// Update the global data.
	members = entries

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

			// Refresh the members list.
			entries, err := parseMembers(fileName)
			if err != nil {
				continue
			}

			members = entries
		}
	}()

	err = watcher.Add(filepath.Dir(fileName))
	if err != nil {
		return err
	}

	return nil
}

func parseMembers(fileName string) (map[string]string, error) {
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
