package glue

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNotFound = errors.New("not found")
)

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

func (s *Store) migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS glue(
        whatsapp_conversation TEXT NOT NULL UNIQUE, -- can refer to a group or DM
        signal_group_id TEXT NOT NULL UNIQUE,
		signal_internal_id TEXT NOT NULL UNIQUE,
		PRIMARY KEY(signal_group_id, signal_internal_id, whatsapp_conversation)
    );
    `

	_, err := s.db.Exec(query)
	return err
}

func (s *Store) GetSignalGroupId(whatsappConversation string) (string, error) {
	query := `
	SELECT signal_group_id
	FROM glue
	WHERE whatsapp_conversation = ?
	`
	var signalGroupId string
	err := s.db.QueryRow(query, whatsappConversation).Scan(&signalGroupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return signalGroupId, nil
}

func (s *Store) GetWhatsAppConversationId(signalInternalId string) (string, error) {
	query := `
	SELECT whatsapp_conversation
	FROM glue
	WHERE signal_internal_id = ?
	`
	var whatsappGroupId string
	err := s.db.QueryRow(query, signalInternalId).Scan(&whatsappGroupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return whatsappGroupId, nil
}
func (s *Store) LinkGroups(whatsAppConversation, signalGroupId, signalInternalId string) error {
	query := `
	INSERT INTO glue(whatsapp_conversation, signal_group_id, signal_internal_id)
	VALUES(?, ?, ?)
	`
	_, err := s.db.Exec(query, whatsAppConversation, signalGroupId, signalInternalId)
	return err
}
