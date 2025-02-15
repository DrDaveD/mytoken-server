package tree

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/oidc-mytoken/api/v0"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/shared/mytoken/pkg/mtid"
	"github.com/oidc-mytoken/server/shared/utils/unixtime"
)

// MytokenEntry holds the information of a MytokenEntry as stored in the
// database
type MytokenEntry struct {
	api.MytokenEntry `json:",inline"`
	ID               mtid.MTID         `json:"-"`
	ParentID         mtid.MTID         `db:"parent_id" json:"-"`
	RootID           mtid.MTID         `db:"root_id" json:"-"`
	Name             db.NullString     `json:"name,omitempty"`
	CreatedAt        unixtime.UnixTime `db:"created" json:"created"`
}

// MytokenEntryTree is a tree of MytokenEntry
type MytokenEntryTree struct {
	Token    MytokenEntry       `json:"token"`
	Children []MytokenEntryTree `json:"children,omitempty"`
}

// Root checks if this MytokenEntry is a root token
func (ste *MytokenEntry) Root() bool {
	if ste.ID.Hash() == ste.RootID.Hash() {
		return true
	}
	return !ste.RootID.HashValid()
}

// getUserID returns the user id linked to a mytoken
func getUserID(tx *sqlx.Tx, tokenID mtid.MTID) (uid int64, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Get(&uid, `SELECT user_id FROM MTokens WHERE id=? ORDER BY name`, tokenID))
	})
	return
}

// AllTokens returns information about all mytokens for the user linked to the passed mytoken
func AllTokens(tx *sqlx.Tx, tokenID mtid.MTID) (trees []MytokenEntryTree, err error) {
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		uid, e := getUserID(tx, tokenID)
		if e != nil {
			return e
		}
		trees, err = allTokensForUser(tx, uid)
		return err
	})
	return
}

// allTokensForUser returns information about all mytoken for the passed user
func allTokensForUser(tx *sqlx.Tx, uid int64) ([]MytokenEntryTree, error) {
	var tokens []MytokenEntry
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Select(&tokens,
			`SELECT id, parent_id, root_id, name, created, ip_created AS ip FROM MTokens WHERE user_id=?`,
			uid))
	}); err != nil {
		return nil, err
	}
	return tokensToTrees(tokens), nil
}

func subtokens(tx *sqlx.Tx, rootID mtid.MTID) ([]MytokenEntry, error) {
	var tokens []MytokenEntry
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		return errors.WithStack(tx.Select(&tokens,
			`SELECT id, parent_id, root_id, name, created, ip_created AS ip FROM MTokens WHERE root_id=?`,
			rootID))
	})
	return tokens, err
}

// TokenSubTree returns information about all subtokens for the passed mytoken
func TokenSubTree(tx *sqlx.Tx, tokenID mtid.MTID) (MytokenEntryTree, error) {
	var tokens []MytokenEntry
	var root MytokenEntry
	if err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		var err error
		if err = tx.Get(&root,
			`SELECT id, parent_id, root_id, name, created, ip_created AS ip FROM MTokens WHERE id=?`,
			tokenID); err != nil {
			return errors.WithStack(err)
		}
		if root.Root() {
			root.RootID = root.ID
		}
		tokens, err = subtokens(tx, root.RootID)
		return err
	}); err != nil {
		return MytokenEntryTree{}, err
	}
	tree, _ := tokensToTree(root, tokens)
	return tree, nil
}

func tokensToTrees(tokens []MytokenEntry) (trees []MytokenEntryTree) {
	var roots []MytokenEntry
	for i := 0; i < len(tokens); {
		t := tokens[i]
		if t.Root() {
			removeEntry(&tokens, i)
			roots = append(roots, t)
		} else {
			i++
		}
	}
	var tmp MytokenEntryTree
	for _, r := range roots {
		tmp, tokens = tokensToTree(r, tokens)
		trees = append(trees, tmp)
	}
	return
}

func tokensToTree(root MytokenEntry, tokens []MytokenEntry) (MytokenEntryTree, []MytokenEntry) {
	tree := MytokenEntryTree{
		Token: root,
	}
	children := popChildren(root, &tokens)
	var cTree MytokenEntryTree
	for _, c := range children {
		cTree, tokens = tokensToTree(c, tokens)
		tree.Children = append(tree.Children, cTree)
	}
	return tree, tokens
}

func popChildren(root MytokenEntry, tokens *[]MytokenEntry) (children []MytokenEntry) {
	i := 0
	for i < len(*tokens) {
		t := (*tokens)[i]
		if t.ParentID == root.ID {
			removeEntry(tokens, i)
			children = append(children, t)
		} else {
			i++
		}
	}
	return
}

func removeEntry(tokens *[]MytokenEntry, i int) { // skipcq SCC-U1000
	copy((*tokens)[i:], (*tokens)[i+1:]) // Shift r[i+1:] left one index.
	*tokens = (*tokens)[:len(*tokens)-1] // Truncate slice.
}
