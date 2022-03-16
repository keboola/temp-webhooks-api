package client

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
)

const LoggerPrefix = "HTTP%s\t"

// Logger for HTTP client.
type Logger struct {
	logger log.Logger
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logWithoutSecretsf("", format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logWithoutSecretsf("-WARN", format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logWithoutSecretsf("-ERROR", format, v...)
}

func (l *Logger) logWithoutSecretsf(level string, format string, v ...interface{}) {
	v = append([]interface{}{level}, v...)
	msg := fmt.Sprintf(LoggerPrefix+format, v...)
	msg = removeSecrets(msg)
	msg = strings.TrimSuffix(msg, "\n")
	l.logger.Debug(msg)
}

func removeSecrets(str string) string {
	return regexp.MustCompile(`(?i)(token[^\w/,]\s*)\d[^\s/]*`).ReplaceAllString(str, "$1*****")
}
