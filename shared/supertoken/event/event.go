package event

import (
	"github.com/jmoiron/sqlx"

	"github.com/oidc-mytoken/server/internal/db/dbrepo/eventrepo"
	"github.com/oidc-mytoken/server/internal/model"
	pkg "github.com/oidc-mytoken/server/shared/supertoken/event/pkg"
	"github.com/oidc-mytoken/server/shared/supertoken/pkg/stid"
)

type MTEvent struct {
	*pkg.Event
	MTID stid.STID
}

// LogEvent logs an event to the database
func LogEvent(tx *sqlx.Tx, event MTEvent, clientMetaData model.ClientMetaData) error {
	return (&eventrepo.EventDBObject{
		Event:          event.Event,
		STID:           event.MTID,
		ClientMetaData: clientMetaData,
	}).Store(tx)
}

// LogEvents logs multiple events for the same token to the database
func LogEvents(tx *sqlx.Tx, events []MTEvent, clientMetaData model.ClientMetaData) error {
	for _, e := range events {
		if err := LogEvent(tx, e, clientMetaData); err != nil {
			return err
		}
	}
	return nil
}
