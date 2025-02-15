package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
	"github.com/oidc-mytoken/server/internal/db/cluster"
)

var db *cluster.Cluster

// Connect connects to the database using the mytoken config
func Connect() {
	ConnectConfig(config.Get().DB)
}

// ConnectConfig connects to the database using the passed config
func ConnectConfig(conf config.DBConf) {
	if db != nil {
		log.Debug("Closing existing db connections")
		db.Close()
	}
	db = cluster.NewFromConfig(conf)
}

// NullString extends the sql.NullString
type NullString struct {
	sql.NullString
}

// MarshalJSON implements the json.Marshaler interface
func (s NullString) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(s.String)
	return data, errors.WithStack(err)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (s *NullString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errors.WithStack(err)
	}
	*s = NewNullString(str)
	return nil
}

// NewNullString creates a new NullString from the given string
func NewNullString(s string) NullString {
	if s == "" {
		return NullString{}
	}
	return NullString{
		sql.NullString{
			String: s,
			Valid:  true,
		},
	}
}

// BitBool is an implementation of a bool for the MySQL type BIT(1).
// This type allows you to avoid wasting an entire byte for MySQL's boolean type TINYINT.
type BitBool bool

// Value implements the driver.Valuer interface,
// and turns the BitBool into a bitfield (BIT(1)) for MySQL storage.
func (b BitBool) Value() (driver.Value, error) {
	if b {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

// Scan implements the sql.Scanner interface,
// and turns the bitfield incoming from MySQL into a BitBool
func (b *BitBool) Scan(src interface{}) error {
	if src == nil {
		*b = false
		return nil
	}
	v, ok := src.([]byte)
	if !ok {
		return errors.New("bad []byte type assertion")
	}
	*b = v[0] == 1
	return nil
}

// Transact does a database transaction for the passed function
func Transact(fn func(*sqlx.Tx) error) error {
	return db.Transact(fn)
}

// RunWithinTransaction runs the passed function using the passed transaction; if nil is passed as tx a new transaction
// is created. This is basically a wrapper function, that works with a possible nil-tx
func RunWithinTransaction(tx *sqlx.Tx, fn func(*sqlx.Tx) error) error {
	return db.RunWithinTransaction(tx, fn)
}
