package configmigrate_test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardDNSClient/internal/configmigrate"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/AdguardTeam/golibs/testutil/faketime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTimeout is the common timeout for tests and contexts.
const testTimeout = 1 * time.Second

// testdataName is the name of the test data directory.
const testdataName = "testdata"

// copyFile is a helper that copies the file from srcPath to dstPath.  srcPath
// must differ from dstPath, and dstPath must not exist.
func copyFile(tb testing.TB, srcPath, dstPath string) {
	tb.Helper()

	require.NotEqual(tb, srcPath, dstPath)
	require.NoFileExists(tb, dstPath)

	srcFile, err := os.Open(srcPath)
	require.NoError(tb, err)
	defer func() { require.NoError(tb, srcFile.Close()) }()

	info, err := srcFile.Stat()
	require.NoError(tb, err)

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, info.Mode())
	require.NoError(tb, err)
	defer func() { require.NoError(tb, dstFile.Close()) }()

	n, err := io.Copy(dstFile, srcFile)
	require.NoError(tb, err)
	require.Equal(tb, info.Size(), n)
}

// assertEqualYAML asserts that the content of the files want and got are equal
// interpreted as YAML documents.
func assertEqualYAML(tb testing.TB, wantPath, gotPath string) (ok bool) {
	tb.Helper()

	wantBytes, err := os.ReadFile(wantPath)
	require.NoError(tb, err)

	gotBytes, err := os.ReadFile(gotPath)
	require.NoError(tb, err)

	return assert.YAMLEq(tb, string(wantBytes), string(gotBytes))
}

// TestMigrator_Run_success tests the successful incremental migrations of the
// configuration files.  The test cases must have the following structure:
//
//	testdata/
//	└── TestMigrator_Run_success/
//	    ├── v<target_schema_version>/
//	    │   ├── in.yaml
//	    │   └── want.yaml
//	    ...
//
// TODO(e.burkov):  Store the input and expected output in a single txtar file,
// allowing to test the same migration properly.
func TestMigrator_Run_success(t *testing.T) {
	t.Parallel()

	const (
		// caseFormat is the format of name for test case directories.
		caseFormat = "v%d"

		inFilename   = "in.yaml"
		wantFilename = "want.yaml"
	)

	testCases, err := fs.Glob(os.DirFS(testdataName), path.Join(t.Name(), "*"))
	require.NoError(t, err)

	curTime := time.Now()
	clock := &faketime.Clock{
		OnNow: func() (now time.Time) { return curTime },
	}
	curDate := curTime.Format(configmigrate.BackupDateTimeFormat)

	tempDir := t.TempDir()

	for _, tc := range testCases {
		testCaseName := filepath.Base(tc)

		var wantVer configmigrate.SchemaVersion
		_, err = fmt.Sscanf(testCaseName, caseFormat, &wantVer)
		require.NoError(t, err)

		var (
			backupDir = fmt.Sprintf(configmigrate.BackupDirNameFormat, wantVer-1, wantVer, curDate)

			configFilename = testCaseName + ".yaml"

			inPath     = filepath.Join(testdataName, tc, inFilename)
			wantPath   = filepath.Join(testdataName, tc, wantFilename)
			outPath    = filepath.Join(tempDir, configFilename)
			backupPath = filepath.Join(tempDir, backupDir, configFilename)
		)

		copyFile(t, inPath, outPath)

		t.Run(testCaseName, func(t *testing.T) {
			t.Parallel()

			migrator := configmigrate.New(&configmigrate.Config{
				Clock:          clock,
				Logger:         slogutil.NewDiscardLogger(),
				WorkingDir:     tempDir,
				ConfigFileName: configFilename,
			})

			ctx := testutil.ContextWithTimeout(t, testTimeout)
			migErr := migrator.Run(ctx, wantVer)
			require.NoError(t, migErr)

			assertEqualYAML(t, wantPath, outPath)
			assertEqualYAML(t, inPath, backupPath)
		})
	}
}
