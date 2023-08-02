package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

var members map[string]string
var overrides map[string]string

func loadMembers(fileName string) error {
	return loadFile(fileName, &members)
}

func loadOverrides(fileName string) error {
	return loadFile(fileName, &overrides)
}

func loadFile(fileName string, target *map[string]string) error {
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

func parseFile(fileName string) (map[string]string, error) {
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

	entries := map[string]string{}
	for _, record := range records {
		entries[record[3]] = fmt.Sprintf("%s (%s)", record[2], record[0])
	}

	return entries, nil
}
