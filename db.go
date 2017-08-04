package main

import (
	"errors"
	"os"
	"sync"

	"github.com/spf13/viper"
	"github.com/tobischo/gokeepasslib"
)

var sharedRoot *gokeepasslib.RootData
var sharedRootLock sync.RWMutex

func GetUserName(entry *gokeepasslib.Entry) string {
	return entry.Get("UserName").Value.Content
}

func GetURL(entry *gokeepasslib.Entry) string {
	return entry.Get("URL").Value.Content
}

func findInGroupByValues(group *gokeepasslib.Group, values map[string]string) (*gokeepasslib.Entry, error) {
	for _, entry := range group.Entries {
		match := true

		for key, value := range values {
			match = match && entry.Get(key).Value.Content == value
		}

		if match {
			return &entry, nil
		}
	}

	for _, innerGroup := range group.Groups {
		entry, err := findInGroupByValues(&innerGroup, values)

		if err == nil {
			return entry, err
		}
	}

	return nil, errors.New("Entry not found")
}

func findInRootByValues(root *gokeepasslib.RootData, values map[string]string) (*gokeepasslib.Entry, error) {
	sharedRootLock.RLock()
	defer sharedRootLock.RUnlock()

	for _, innerGroup := range root.Groups {
		entry, err := findInGroupByValues(&innerGroup, values)

		if err == nil {
			return entry, err
		}
	}

	return nil, errors.New("Entry not found")
}

func loadDB(password string) error {
	path := viper.GetString("keepass-file")
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return err
	}

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	err = gokeepasslib.NewDecoder(file).Decode(db)

	if err != nil {
		return err
	}

	db.UnlockProtectedEntries()

	sharedRootLock.Lock()
	defer sharedRootLock.Unlock()
	sharedRoot = db.Content.Root

	return nil
}
