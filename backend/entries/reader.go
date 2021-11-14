package entries

import (
	"github.com/mtlynch/whatgotdone/backend/datastore"
	"github.com/mtlynch/whatgotdone/backend/types"
)

// Reader reads journal entries.
type Reader interface {
	// Recent returns the recent entries in the store.
	Recent(start, limit int) ([]types.JournalEntry, error)
	// RecentFollowing returns recent entries from among users that the specified
	// user is following.
	RecentFollowing(username types.Username, start, limit int) ([]types.JournalEntry, error)
}

// EntryStore stores information related to journal entries.
type EntryStore interface {
	// Users returns all the users who have published entries.
	Users() ([]types.Username, error)
	// ReadEntries returns all published entries matching the given filter.
	ReadEntries(filter datastore.EntryFilter) ([]types.JournalEntry, error)
	// GetReactions retrieves reader reactions associated with a published entry.
	// Followers returns all the users the specified user is following.
	Following(follower types.Username) ([]types.Username, error)
}

type defaultReader struct {
	store EntryStore
}

// NewReader creates a new entries.Reader.
func NewReader(store EntryStore) Reader {
	return defaultReader{
		store: store,
	}
}
