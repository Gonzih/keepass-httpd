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
	"github.com/tobischo/gokeepasslib"
)

type responseEntry struct {
	UserName string `json:"username"`
	Title    string `json:"title"`
	Password string `json:"password"`
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

func findInGroupByUserName(group *gokeepasslib.Group, userName string) (*gokeepasslib.Entry, error) {
	for _, entry := range group.Entries {
		if userName == GetUserName(&entry) {
			return &entry, nil
		}
	}

	for _, innerGroup := range group.Groups {
		entry, err := findInGroupByUserName(&innerGroup, userName)

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
		Password: entry.GetPassword()}

	return json.Marshal(&response)
}

var sharedGroup gokeepasslib.Group
var sharedGroupLock sync.RWMutex

func loadDB() error {
	file, _ := os.Open("test.kdbx")
	defer file.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials("test")
	err := gokeepasslib.NewDecoder(file).Decode(db)

	if err != nil {
		return err
	}

	db.UnlockProtectedEntries()

	sharedGroupLock.Lock()
	defer sharedGroupLock.Unlock()
	sharedGroup = db.Content.Root.Groups[0]

	return nil
}

func init() {
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

func findEntryByUsername(username string) (*gokeepasslib.Entry, error) {
	sharedGroupLock.RLock()
	defer sharedGroupLock.RUnlock()
	return findInGroupByUserName(&sharedGroup, username)
}

func SearchHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	username := r.FormValue("username")

	if len(username) == 0 {
		respondWithError(w, errors.New("username parameter is required"), http.StatusBadRequest)
		return
	}

	entry, err := findEntryByUsername(username)

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
