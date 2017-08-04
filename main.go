package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tobischo/gokeepasslib"
)

type responseEntry struct {
	UserName string `json:"username"`
	Title    string `json:"title"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	Status string `json:"status"`
}

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

func marshalEntry(entry *gokeepasslib.Entry) ([]byte, error) {
	response := responseEntry{
		UserName: GetUserName(entry),
		Title:    entry.GetTitle(),
		Password: entry.GetPassword(),
		URL:      GetURL(entry)}

	return json.Marshal(&response)
}

var sharedGroup gokeepasslib.Group
var sharedGroupLock sync.RWMutex

func loadDB() error {
	path := viper.GetString("keepass-file")
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return fmt.Errorf("Error while opening \"%s\": %s", path, err)
	}

	db := gokeepasslib.NewDatabase()
	pass := viper.GetString("keepass-password")
	db.Credentials = gokeepasslib.NewPasswordCredentials(pass)
	err = gokeepasslib.NewDecoder(file).Decode(db)

	if err != nil {
		return err
	}

	db.UnlockProtectedEntries()

	sharedGroupLock.Lock()
	defer sharedGroupLock.Unlock()
	sharedGroup = db.Content.Root.Groups[0]

	return nil
}

func initViper() {
	pflag.String("keepass-password", "", "KeepassDB password")
	pflag.String("keepass-file", "", "KeepassDB file path")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}

func init() {
	initViper()

	err := loadDB()

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	router := httprouter.New()

	router.GET("/search", SearchHandler)
	router.POST("/reload", ReloadHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func respondWithError(w http.ResponseWriter, err error, status int) {
	response := errorResponse{Error: err.Error()}
	json, err := json.Marshal(&response)

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func findEntry(values map[string]string) (*gokeepasslib.Entry, error) {
	sharedGroupLock.RLock()
	defer sharedGroupLock.RUnlock()
	return findInGroupByValues(&sharedGroup, values)
}

func SearchHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := r.FormValue("username")
	title := r.FormValue("title")
	url := r.FormValue("url")

	if len(username) == 0 && len(title) == 0 && len(url) == 0 {
		respondWithError(w, errors.New("from username/title/url at least one parameter is required"), http.StatusBadRequest)
		return
	}

	searchValues := make(map[string]string)

	if len(username) > 0 {
		searchValues["UserName"] = username
	}
	if len(title) > 0 {
		searchValues["Title"] = title
	}
	if len(url) > 0 {
		searchValues["URL"] = url
	}

	entry, err := findEntry(searchValues)

	if err != nil {
		respondWithError(w, err, http.StatusNotFound)
		return
	}

	json, err := marshalEntry(entry)

	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func ReloadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := loadDB()

	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	response := successResponse{Status: "success"}
	json, err := json.Marshal(&response)

	if err != nil {
		respondWithError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
