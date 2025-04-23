package configmigrate

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/agdcos"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/AdguardTeam/golibs/validate"
	"github.com/google/renameio/v2/maybe"
	"gopkg.in/yaml.v3"
)

// Config is the configuration for [Migrator].
type Config struct {
	// Clock is the source of current time.
	Clock timeutil.Clock

	// Logger used to log migrator operations.
	Logger *slog.Logger

	// WorkingDir is the absolute path to the working directory of AdGuardDNS
	// Client.
	WorkingDir string

	// ConfigFileName is the name of the configuration file within the working
	// directory.
	ConfigFileName string
}

// Migrator performs the YAML configuration file migrations.
type Migrator struct {
	clock      timeutil.Clock
	logger     *slog.Logger
	workingDir string
	configName string
}

// New creates a new Migrator.
func New(c *Config) (m *Migrator) {
	return &Migrator{
		clock:      c.Clock,
		logger:     c.Logger,
		workingDir: c.WorkingDir,
		configName: c.ConfigFileName,
	}
}

// Run performs necessary upgrade operations to upgrade file to target schema
// version, if needed.
func (m *Migrator) Run(ctx context.Context, target SchemaVersion) (err error) {
	defer func() { err = errors.Annotate(err, "migrating: %w") }()

	confPath := filepath.Join(m.workingDir, m.configName)
	logger := m.logger.With("config_path", confPath)

	conf, confData, err := readYAML(confPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			logger.DebugContext(ctx, "configuration file does not exist")

			return nil
		}

		// Don't wrap the error since it's informative enough as is.
		return err
	}

	current, err := m.getVersion(conf)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return err
	} else if current == target {
		logger.DebugContext(ctx, "configuration file needs no migration", "version", current)

		return nil
	}

	logger.DebugContext(ctx, "migrating configuration file")

	err = m.migrate(ctx, conf, current, target)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return err
	}

	// Don't wrap the error since it's informative enough as is.
	return m.writeMigrated(ctx, logger, conf, confData, confPath, current, target)
}

// getVersion returns the schema version from the configuration object.
func (m *Migrator) getVersion(conf yObj) (v SchemaVersion, err error) {
	verInt, err := fieldVal[int](conf, SchemaVersionKey)
	if err != nil {
		return 0, err
	}

	err = validate.InRange(SchemaVersionKey, verInt, int(VersionInitial), int(VersionLatest))
	if err != nil {
		return 0, err
	}

	// #nosec G115 -- The value is validated to be within the range.
	return SchemaVersion(verInt), nil
}

// migrateFunc is a function that upgrades conf or returns an error.
type migrateFunc = func(ctx context.Context, conf yObj) (err error)

// migrate performs migrations from current schema version to the target one.
// It returns an error from the first failed migration.
func (m *Migrator) migrate(ctx context.Context, conf yObj, curr, targ SchemaVersion) (err error) {
	migrations := [VersionLatest]migrateFunc{
		// There is obviously no migration to the initial version.
		0: nil,
		1: m.migrateTo2,
		2: m.migrateTo3,
	}

	for i, migrate := range migrations[curr:targ] {
		// #nosec G115 -- The value is guaranteed to be less than the upper
		// bound of the array.
		cur := curr + SchemaVersion(i)
		next := cur + 1

		m.logger.InfoContext(ctx, "upgrading configuration", "from", cur, "to", next)

		err = migrate(ctx, conf)
		if err != nil {
			return fmt.Errorf("upgrading schema %d to %d: %w", cur, next, err)
		}
	}

	return nil
}

// BackupDateTimeFormat is the format for timestamps in the backup directory
// name.
const BackupDateTimeFormat = time.DateOnly + "-15-04-05"

// BackupDirNameFormat is the format for the backup directory name, which should
// be filled with the current and target schema versions, and current datetime
// formatted according to [BackupDateTimeFormat].
const BackupDirNameFormat = "backup-configmigrate-v%d-to-v%d-%s.dir"

// backupConfig creates a backupConfig of the original configuration file,
// properly naming it to avoid conflicts.
func (m *Migrator) backupConfig(
	ctx context.Context,
	origData []byte,
	curr SchemaVersion,
	targ SchemaVersion,
) (err error) {
	// Create a backup directory if it doesn't exist.
	bkpTime := m.clock.Now().Format(BackupDateTimeFormat)
	bkpDirName := fmt.Sprintf(BackupDirNameFormat, curr, targ, bkpTime)
	bkpDirPath := filepath.Join(m.workingDir, bkpDirName)

	// Don't use [os.MkdirAll] since the backup directory is expected to be
	// right in the working directory and should not yet exist.
	err = os.Mkdir(bkpDirPath, agdcos.DefaultPermDir)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return fmt.Errorf("creating backup directory: %w", err)
	}

	bkpPath := filepath.Join(bkpDirPath, m.configName)

	err = maybe.WriteFile(bkpPath, origData, agdcos.DefaultPermFile)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return fmt.Errorf("moving original configuration file: %w", err)
	}

	m.logger.DebugContext(ctx, "configuration file backed up", "backup_path", bkpPath)

	return nil
}

// writeMigrated writes the migrated configuration to the original configuration
// path, moving the file itself to backup path.
func (m *Migrator) writeMigrated(
	ctx context.Context,
	l *slog.Logger,
	conf yObj,
	origData []byte,
	origPath string,
	curr SchemaVersion,
	targ SchemaVersion,
) (err error) {
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)

	err = enc.Encode(conf)
	if err != nil {
		return fmt.Errorf("encoding migrated configuration: %w", err)
	}

	err = enc.Close()
	if err != nil {
		return fmt.Errorf("closing the encoder: %w", err)
	}

	err = m.backupConfig(ctx, origData, curr, targ)
	if err != nil {
		return fmt.Errorf("creating backup of configuration file: %w", err)
	}

	l.DebugContext(ctx, "writing migrated configuration")

	// TODO(e.burkov):  Take care of permissions on Windows.

	err = maybe.WriteFile(origPath, buf.Bytes(), agdcos.DefaultPermFile)
	if err != nil {
		return fmt.Errorf("writing migrated configuration: %w", err)
	}

	l.DebugContext(ctx, "migrated configuration written")

	return nil
}
