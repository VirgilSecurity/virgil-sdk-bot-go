package webhook

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/VirgilSecurity/virgil-sdk-bot-go/storage"
)

func TestClient(t *testing.T) {
	strg := &storage.FileStorage{}
	cli, err := NewClient("https://api-dev.virgilsecurity.com/f4f4f4f4f4f4f4f4/f4f4f4f4f4f4f4f4f4f4f4f4f4f4f4f4", strg)
	require.NoError(t, err)
	require.NoError(t, cli.Init())
}
