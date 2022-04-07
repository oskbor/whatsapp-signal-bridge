package glue_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/oskbor/bridge/glue"
	"github.com/stretchr/testify/require"
)

var store *glue.Store

func TestMain(m *testing.M) {
	db, err := sql.Open("sqlite3", "file:glue_test.db?_foreign_keys=on")
	if err != nil {
		panic(err)
	}

	store, err = glue.NewStore(db)
	if err != nil {
		panic(err)
	}
	code := m.Run()
	db.Close()
	err = os.Remove("glue_test.db")
	if err != nil {
		panic(err)
	}

	os.Exit(code)
}

func TestDbBasics(t *testing.T) {
	t.Run("Link two conversations and query from both sides", func(t *testing.T) {
		waNumber := "+123"
		signalNumber := "+456"
		waConversation := "group1"
		signalGroup := "group2"
		err := store.LinkGroups(waConversation, waNumber, signalGroup, signalNumber)
		require.Nil(t, err)
		waGroupId, err := store.GetWhatsAppConversationId(signalGroup, signalNumber, waNumber)
		require.Nil(t, err)
		require.Equal(t, waGroupId, waConversation)
		signalGroupId, err := store.GetSignalGroupId(waConversation, waNumber, signalNumber)
		require.Nil(t, err)
		require.Equal(t, signalGroupId, signalGroup)
	})

}
