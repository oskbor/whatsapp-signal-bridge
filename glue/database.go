package glue

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

/*
var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)
*/
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) (*Store, error) {
	s := &Store{
		db: db,
	}
	err := s.migrate()
	if err != nil {
		return nil, err
	}
	return s, err
}

func (r *Store) migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS glue(
        whatsapp_conversation TEXT NOT NULL, -- can refer to a group or DM
		whatsapp_phonenumber TEXT NOT NULL,  
        signal_group TEXT NOT NULL UNIQUE,
		signal_phonenumber TEXT NOT NULL,
		PRIMARY KEY(signal_phonenumber, whatsapp_phonenumber, whatsapp_conversation)
    );
    `

	_, err := r.db.Exec(query)
	return err
}
