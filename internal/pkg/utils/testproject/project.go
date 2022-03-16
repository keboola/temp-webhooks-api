// nolint forbidigo
package testproject

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/juju/fslock"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"github.com/keboola/temp-webhooks-api/internal/pkg/api/storageapi"
	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/utils/testapi"
	"github.com/keboola/temp-webhooks-api/internal/pkg/utils/testhelper"
)

type Project struct {
	t          *testing.T
	host       string // Storage API host
	token      string // Storage API token
	id         int    // project ID
	lock       *fslock.Lock
	locked     bool
	mutex      *sync.Mutex
	storageApi *storageapi.Api
	envLock    *sync.Mutex
	envs       *env.Map
	newEnvs    []string
}

// newProject - create test project handler and lock it.
func newProject(host string, id int, token string) *Project {
	// Create locks dir if not exists
	locksDir := filepath.Join(os.TempDir(), `.keboola-as-code-locks`)
	if err := os.MkdirAll(locksDir, 0o700); err != nil {
		panic(fmt.Errorf(`cannot lock test project: %s`, err))
	}

	// lock file name
	lockFile := host + `-` + cast.ToString(id) + `.lock`
	lockPath := filepath.Join(locksDir, lockFile)

	p := &Project{
		host:    host,
		id:      id,
		token:   token,
		lock:    fslock.New(lockPath),
		mutex:   &sync.Mutex{},
		envLock: &sync.Mutex{},
	}

	// Init API
	p.storageApi, _ = testapi.NewStorageApiWithToken(p.host, p.token, testhelper.TestIsVerbose())

	logger := log.NewDebugLogger()
	if testhelper.TestIsVerbose() {
		logger.ConnectTo(os.Stdout)
	}

	// Check project ID
	if p.id != p.storageApi.ProjectId() {
		assert.FailNow(p.t, "test project id and token project id are different.")
	}

	return p
}

func (p *Project) Id() int {
	p.assertLocked()
	return p.id
}

func (p *Project) Name() string {
	p.assertLocked()
	return p.storageApi.ProjectName()
}

func (p *Project) StorageApiHost() string {
	p.assertLocked()
	return p.host
}

func (p *Project) StorageApiToken() string {
	p.assertLocked()
	return p.storageApi.Token().Token
}

func (p *Project) StorageApi() *storageapi.Api {
	p.assertLocked()
	return p.storageApi
}

// setEnv set ENV variable, all ENVs are logged at the end of SetState method.
func (p *Project) setEnv(key string, value string) {
	// Normalize key
	key = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(key, "_")
	key = strings.ToUpper(key)
	key = strings.Trim(key, "_")

	// Set
	p.envs.Set(key, value)

	// Log
	p.envLock.Lock()
	defer p.envLock.Unlock()
	p.newEnvs = append(p.newEnvs, fmt.Sprintf("%s=%s", key, value))
}

func (p *Project) logEnvs() {
	for _, item := range p.newEnvs {
		p.logf(fmt.Sprintf(`ENV "%s"`, item))
	}
}

func (p *Project) logf(format string, a ...interface{}) {
	if testhelper.TestIsVerbose() {
		a = append([]interface{}{p.id, p.t.Name()}, a...)
		p.t.Logf("TestProject[%d][%s]: "+format, a...)
	}
}

func (p *Project) assertLocked() {
	if !p.locked {
		panic(fmt.Errorf(`test project "%d" is not locked`, p.id))
	}
}

func (p *Project) tryLock(t *testing.T, envs *env.Map) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.locked {
		return false
	}

	if err := p.lock.TryLock(); err != nil {
		if !errors.Is(err, fslock.ErrLocked) {
			// Unexpected error
			panic(err)
		}

		// Busy
		return false
	}

	// Locked!
	p.t = t
	p.locked = true

	// Unlock, when test is done
	p.t.Cleanup(func() {
		p.unlock()
	})

	// Set ENVs, the environment resets when unlock is called
	p.envs = envs
	p.newEnvs = make([]string, 0)
	p.setEnv(`TEST_KBC_PROJECT_ID`, cast.ToString(p.Id()))
	p.setEnv(`TEST_KBC_PROJECT_NAME`, p.Name())
	p.setEnv(`TEST_KBC_STORAGE_API_HOST`, p.StorageApiHost())
	p.setEnv(`TEST_KBC_STORAGE_API_TOKEN`, p.StorageApiToken())
	p.logf(`Project locked`)

	return true
}

// unlock project if it is no more needed in test.
func (p *Project) unlock() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.newEnvs = make([]string, 0)
	p.envs = nil
	p.locked = false
	p.logf(`Project unlocked`)
	p.t = nil

	if err := p.lock.Unlock(); err != nil {
		panic(fmt.Errorf(`cannot unlock test project: %w`, err))
	}
}
