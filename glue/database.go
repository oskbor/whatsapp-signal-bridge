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
        whatsapp_conversation TEXT NOT NULL, -- can refer to a group or DM
		whatsapp_phonenumber TEXT NOT NULL,  
        signal_group TEXT NOT NULL UNIQUE,
		signal_phonenumber TEXT NOT NULL,
		PRIMARY KEY(signal_phonenumber, whatsapp_phonenumber, whatsapp_conversation)
    );
    `

	_, err := s.db.Exec(query)
	return err
}

func (s *Store) GetSignalGroupId(whatsappConversation, whatsappPhoneNumber, signalPhoneNumber string) (string, error) {
	query := `
	SELECT signal_group
	FROM glue
	WHERE whatsapp_conversation = ?
	AND whatsapp_phonenumber = ?
	AND signal_phonenumber = ?
	`
	var signalGroupId string
	err := s.db.QueryRow(query, whatsappConversation, whatsappPhoneNumber, signalPhoneNumber).Scan(&signalGroupId)
	if err != nil {
		return "", err
	}
	return signalGroupId, nil
}

func (s *Store) GetWhatsAppConversationId(signalGroupId, signalPhoneNumber, whatsappPhoneNumber string) (string, error) {
	query := `
	SELECT whatsapp_conversation
	FROM glue
	WHERE signal_group = ?
	AND signal_phonenumber = ?
	AND whatsapp_phonenumber = ?
	`
	var whatsappGroupId string
	err := s.db.QueryRow(query, signalGroupId, signalPhoneNumber, whatsappPhoneNumber).Scan(&whatsappGroupId)
	if err != nil {
		return "", err
	}
	return whatsappGroupId, nil
}
func (s *Store) LinkGroups(whatsappConversation, whatsappPhoneNumber, signalGroupId, signalPhoneNumber string) error {
	query := `
	INSERT INTO glue(whatsapp_conversation, whatsapp_phonenumber, signal_group, signal_phonenumber)
	VALUES(?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, whatsappConversation, whatsappPhoneNumber, signalGroupId, signalPhoneNumber)
	return err
}
