package services

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var services Service

func TestMain(m *testing.M) {

	services = NewService()
	os.Exit(m.Run())
}
func TestSomeService(t *testing.T) {
	doc, err := services.SomeService()
	require.NoError(t, err)
	require.NotEmpty(t, doc)
}
