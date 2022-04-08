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

func (s *Store) migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS glue(
        whatsapp_conversation TEXT NOT NULL UNIQUE, -- can refer to a group or DM
        signal_group TEXT NOT NULL UNIQUE,
		PRIMARY KEY(signal_group, whatsapp_conversation)
    );
    `

	_, err := s.db.Exec(query)
	return err
}

func (s *Store) GetSignalGroupId(whatsappConversation string) (string, error) {
	query := `
	SELECT signal_group
	FROM glue
	WHERE whatsapp_conversation = ?
	`
	var signalGroupId string
	err := s.db.QueryRow(query, whatsappConversation).Scan(&signalGroupId)
	if err != nil {
		return "", err
	}
	return signalGroupId, nil
}

func (s *Store) GetWhatsAppConversationId(signalGroupId string) (string, error) {
	query := `
	SELECT whatsapp_conversation
	FROM glue
	WHERE signal_group = ?
	`
	var whatsappGroupId string
	err := s.db.QueryRow(query, signalGroupId).Scan(&whatsappGroupId)
	if err != nil {
		return "", err
	}
	return whatsappGroupId, nil
}
func (s *Store) LinkGroups(whatsappConversation, signalGroupId string) error {
	query := `
	INSERT INTO glue(whatsapp_conversation, signal_group)
	VALUES(?, ?)
	`
	_, err := s.db.Exec(query, whatsappConversation, signalGroupId)
	return err
}
