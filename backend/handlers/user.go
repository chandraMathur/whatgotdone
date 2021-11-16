package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mtlynch/whatgotdone/backend/datastore"
	"github.com/mtlynch/whatgotdone/backend/handlers/parse"
	"github.com/mtlynch/whatgotdone/backend/types"
	"github.com/mtlynch/whatgotdone/backend/types/requests"
)

func (s defaultServer) userGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := usernameFromRequestPath(r)
		if err != nil {
			log.Printf("Failed to retrieve username from request path: %s", err)
			http.Error(w, "Invalid username", http.StatusBadRequest)
			return
		}

		p, err := s.datastore.GetUserProfile(username)
		if _, ok := err.(datastore.UserProfileNotFoundError); ok {
			http.Error(w, "No profile found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Failed to retrieve user profile data for %s: %s", username, err)
			http.Error(w, "Invalid username", http.StatusNotFound)
			return
		}

		respondOK(w, struct {
			AboutMarkdown   string `json:"aboutMarkdown"`
			TwitterHandle   string `json:"twitterHandle"`
			EmailAddress    string `json:"emailAddress"`
			MastodonAddress string `json:"mastodonAddress"`
		}{
			AboutMarkdown:   string(p.AboutMarkdown),
			TwitterHandle:   string(p.TwitterHandle),
			EmailAddress:    string(p.EmailAddress),
			MastodonAddress: string(p.MastodonAddress),
		})
	}
}

func (s defaultServer) userPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := s.loggedInUser(r)
		if err != nil {
			http.Error(w, "You must log in to update your profile", http.StatusForbidden)
			return
		}

		userProfile, err := profileFromRequest(r)
		if err != nil {
			log.Printf("Invalid profile update request: %v", err)
			http.Error(w, fmt.Sprintf("Invalid profile update request: %v", err), http.StatusBadRequest)
			return
		}

		err = s.datastore.SetUserProfile(username, userProfile)
		if err != nil {
			log.Printf("Failed to update user profile: %s", err)
			http.Error(w, "Failed to update user profile", http.StatusInternalServerError)
			return
		}
	}
}

func (s defaultServer) userMeGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, err := s.loggedInUser(r)
		if err != nil {
			http.Error(w, "You must be logged in to retrieve information about your account", http.StatusForbidden)
			return
		}

		respondOK(w, struct {
			Username types.Username `json:"username"`
		}{
			Username: username,
		})
	}
}

func profileFromRequest(r *http.Request) (types.UserProfile, error) {
	var pur requests.ProfileUpdate
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&pur)
	if err != nil {
		return types.UserProfile{}, err
	}
	return parse.ProfileUpdateRequest(pur)
}

func (s defaultServer) userExists(username types.Username) (bool, error) {
	// BUG: Will only detect users who have published an entry. Ideally, we'd be
	// able to tell if the username exists in UserKit, but the UserKit API
	// currently does not offer lookup by username.
	entries, err := s.datastore.ReadEntries(datastore.EntryFilter{
		ByUsers: []types.Username{
			username,
		},
	})
	if err != nil {
		return false, err
	}

	// TODO: Verify that ReadEntries returns empty slice on no results instead of
	// returning an error.
	if len(entries) > 0 {
		return true, nil
	}

	return false, nil
}
