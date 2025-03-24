// Package configmigrate used to migrate older configuration versions to the
// current one.
package configmigrate

// SchemaVersion is the type for the configuration structure revision.
type SchemaVersion uint

const (
	// VersionInitial is the first version of the configuration structure ever
	// existed.
	VersionInitial SchemaVersion = 1

	// VersionLatest is the current version of the configuration structure.
	VersionLatest SchemaVersion = 2
)

// SchemaVersionKey is the key for the schema version in the YAML configuration
// file.
//
// NOTE: Don't change this value.
const SchemaVersionKey = "schema_version"
