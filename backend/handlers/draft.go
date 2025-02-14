package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/whatgotdone/backend/datastore"
	"github.com/mtlynch/whatgotdone/backend/handlers/parse"
	"github.com/mtlynch/whatgotdone/backend/types"
)

func (s defaultServer) draftGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := usernameFromContext(r.Context())

		date, err := dateFromRequestPath(r)
		if err != nil {
			log.Printf("Invalid date: %s", date)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		draftMarkdown, err := s.savedDraftOrEntryTemplate(username, date)
		if err != nil {
			log.Printf("Failed to retrieve draft entry: %s", err)
			http.Error(w, "Failed to retrieve draft entry", http.StatusInternalServerError)
			return
		}
		if draftMarkdown == "" {
			http.Error(w, "No draft found for this entry", http.StatusNotFound)
			return
		}

		respondOK(w, struct {
			Markdown string `json:"markdown"`
		}{
			Markdown: string(draftMarkdown),
		})
	}
}

func (s defaultServer) draftPut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date, err := dateFromRequestPath(r)
		if err != nil {
			log.Printf("Invalid date: %s", date)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		entryContent, err := entryContentFromRequest(r)
		if err != nil {
			log.Printf("Invalid draft request: %v", err)
			http.Error(w, fmt.Sprintf("Invalid draft request: %v", err), http.StatusBadRequest)
			return
		}

		username := usernameFromContext(r.Context())
		err = s.datastore.InsertDraft(username, types.JournalEntry{
			Date:     date,
			Markdown: entryContent,
		})
		if err != nil {
			log.Printf("Failed to update draft entry: %s", err)
			http.Error(w, "Failed to update draft entry", http.StatusInternalServerError)
			return
		}
	}
}

func (s *defaultServer) draftDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date, err := dateFromRequestPath(r)
		if err != nil {
			log.Printf("Invalid date: %s", date)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		username := usernameFromContext(r.Context())

		err = s.datastore.DeleteDraft(username, date)
		if err != nil {
			log.Printf("Failed to delete draft entry: %s", err)
			http.Error(w, "Failed to delete entry", http.StatusInternalServerError)
			return
		}
	}
}

func (s defaultServer) savedDraftOrEntryTemplate(username types.Username, date types.EntryDate) (types.EntryContent, error) {
	// First, check if there's a saved draft.
	d, err := s.datastore.GetDraft(username, date)
	if _, ok := err.(datastore.DraftNotFoundError); ok {
		// If there's no saved draft, try using the user's entry template.
		return s.getEntryTemplate(username, date)
	} else if err != nil {
		return "", err
	}

	return d.Markdown, nil
}

func (s defaultServer) getEntryTemplate(username types.Username, date types.EntryDate) (types.EntryContent, error) {
	prefs, err := s.datastore.GetPreferences(username)
	if _, ok := err.(datastore.PreferencesNotFoundError); ok {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return prefs.EntryTemplate, nil
}

func entryContentFromRequest(r *http.Request) (types.EntryContent, error) {
	cr := struct {
		EntryContent string `json:"entryContent"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&cr)
	if err != nil {
		return types.EntryContent(""), err
	}

	return parse.EntryContent(cr.EntryContent)
}
