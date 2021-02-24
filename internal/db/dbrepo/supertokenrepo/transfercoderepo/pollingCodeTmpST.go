package transfercoderepo

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db"
	"github.com/oidc-mytoken/server/internal/db/dbrepo/authcodeinforepo/state"
	"github.com/oidc-mytoken/server/pkg/model"
)

// TransferCodeStatus holds information about the status of a polling code
type TransferCodeStatus struct {
	Found           bool               `db:"found"`
	Expired         bool               `db:"expired"`
	ResponseType    model.ResponseType `db:"response_type"`
	ConsentDeclined db.BitBool         `db:"consent_declined"`
}

// CheckTransferCode checks the passed polling code in the database
func CheckTransferCode(tx *sqlx.Tx, pollingCode string) (TransferCodeStatus, error) {
	pt := createProxyToken(pollingCode)
	var p TransferCodeStatus
	err := db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		if err := tx.Get(&p, `SELECT 1 as found, CURRENT_TIMESTAMP() > expires_at AS expired, response_type, consent_declined FROM TransferCodes WHERE id=?`, pt.ID()); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil  // polling code was not found, but this is fine
				return err // p.Found is false
			}
			return err
		}
		return nil
	})
	return p, err
}

// PopTokenForTransferCode returns the decrypted token for a polling code and then deletes the entry
func PopTokenForTransferCode(tx *sqlx.Tx, pollingCode string) (jwt string, err error) {
	pt := createProxyToken(pollingCode)
	var valid bool
	err = db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		jwt, valid, err = pt.JWT(tx)
		if err != nil {
			return err
		}
		if !valid || jwt == "" {
			return nil
		}
		return pt.Delete(tx)
	})
	return
}

// LinkPollingCodeToST links a pollingCode to a SuperToken
func LinkPollingCodeToST(tx *sqlx.Tx, pollingCode, jwt string) error {
	pc := createProxyToken(pollingCode)
	if err := pc.SetJWT(jwt); err != nil {
		return err
	}
	return pc.Update(tx)
}

// DeleteTransferCodeByState deletes a polling code
func DeleteTransferCodeByState(tx *sqlx.Tx, state *state.State) error {
	pc := createProxyToken(state.PollingCode())
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`DELETE FROM ProxyTokens WHERE id=?`, pc.ID())
		return err
	})
}

// DeclineConsentByState updates the polling code attribute after the consent has been declined
func DeclineConsentByState(tx *sqlx.Tx, state *state.State) error {
	pc := createProxyToken(state.PollingCode())
	return db.RunWithinTransaction(tx, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`UPDATE TransferCodesAttributes SET consent_declined=? WHERE id=?`, db.BitBool(true), pc.ID())
		return err
	})
}
