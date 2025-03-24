package configmigrate

import (
	"context"
	"fmt"
	"time"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/timeutil"
)

// migrateToV2 migrates the configuration from version 1 to version 2.  It adds
// the bind_retry field to the dns.server section:
//
// # Before:
//
//	dns:
//	    server:
//	        # …
//	    # …
//	# …
//	schema_version: 1
//
// # After:
//
//	dns:
//	    server:
//	        bind_retry:
//	            enabled: true
//	            interval: 1s
//	            count: 4
//	        # …
//	    # …
//	# …
//	schema_version: 2
func (m *Migrator) migrateTo2(ctx context.Context, conf yObj) (err error) {
	const target SchemaVersion = 2

	dnsVal, err := fieldVal[yObj](conf, "dns")
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return err
	}

	serverVal, err := fieldVal[yObj](dnsVal, "server")
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return err
	}

	const key = "bind_retry"

	_, ok := serverVal[key]
	if ok {
		// TODO(e.burkov):  Add errors.ErrNotNil.
		return fmt.Errorf("%s: %w", key, errors.ErrNotEmpty)
	}

	serverVal[key] = yObj{
		"enabled":  true,
		"interval": timeutil.Duration(1 * time.Second),
		"count":    4,
	}

	conf[SchemaVersionKey] = target

	return nil
}
