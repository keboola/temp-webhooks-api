package client

import (
	"testing"

	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestClientLogger(t *testing.T) {
	t.Parallel()
	logger := log.NewDebugLogger()
	clientLogger := &Logger{logger}
	clientLogger.Debugf("Some debug")
	clientLogger.Warnf("Some warning")
	clientLogger.Errorf("Some error")

	expected := "DEBUG  HTTP\tSome debug\nDEBUG  HTTP-WARN\tSome warning\nDEBUG  HTTP-ERROR\tSome error\n"
	assert.Equal(t, expected, logger.AllMessages())
}

func TestRemoveSecrets(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "token/verify", removeSecrets("token/verify"))
	assert.Equal(t, "tokens/verify", removeSecrets("tokens/verify"))
	assert.Equal(t, "token, abc", removeSecrets("token, abc"))
	assert.Equal(t, "Access token must be set", removeSecrets("Access token must be set"))
	assert.Equal(t, "token: *****", removeSecrets("token: 123ABC12345-abc"))
	assert.Equal(t, "token: ***** ", removeSecrets("token: 123ABC12345-abc "))
	assert.Equal(t, "foo1: bar1\ntoken: *****\nfoo2: bar2", removeSecrets("foo1: bar1\ntoken: 123ABC12345-abc\nfoo2: bar2"))
	assert.Equal(t, "X-StorageApi-Token: *****", removeSecrets("X-StorageApi-Token: 123ABC12345-abc"))
}
