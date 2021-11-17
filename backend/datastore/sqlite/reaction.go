package sqlite

import (
	"log"

	"github.com/mtlynch/whatgotdone/backend/types"
)

// GetReactions retrieves reader reactions associated with a published entry.
func (d db) GetReactions(entryAuthor types.Username, entryDate types.EntryDate) ([]types.Reaction, error) {
	stmt, err := d.ctx.Prepare(`
	SELECT
		reacting_user,
		reaction,
		timestamp
	FROM
		entry_reactions
	WHERE
		entry_author=? AND
		entry_date=?`)
	if err != nil {
		return []types.Reaction{}, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(entryAuthor, entryDate)
	if err != nil {
		return []types.Reaction{}, err
	}

	reactions := []types.Reaction{}
	for rows.Next() {
		var user string
		var reaction string
		var timestampRaw string
		err := rows.Scan(&user, &reaction, &timestampRaw)
		if err != nil {
			return []types.Reaction{}, err
		}

		t, err := parseDatetime(timestampRaw)
		if err != nil {
			return []types.Reaction{}, err
		}

		reactions = append(reactions, types.Reaction{
			Username:  types.Username(user),
			Symbol:    reaction,
			Timestamp: t.Format("2006-01-02 15:04:05Z"),
		})
	}

	return reactions, nil
}

// AddReaction saves a reader reaction associated with a published entry,
// overwriting any existing reaction.
func (d db) AddReaction(entryAuthor types.Username, entryDate types.EntryDate, reaction types.Reaction) error {
	log.Printf("saving reaction to datastore: %s to %s/%s: [%s]", reaction.Username, entryAuthor, entryDate, reaction.Symbol)
	_, err := d.ctx.Exec(`
	INSERT OR REPLACE INTO entry_reactions(
		entry_author,
		entry_date,
		reacting_user,
		reaction,
		timestamp)
	values(?,?,?,?,datetime('now'))`, entryAuthor, entryDate, reaction.Username, reaction.Symbol)
	return err
}
