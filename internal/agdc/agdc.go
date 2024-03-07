// Package agdc contains some types and functions commonly used in AdGuardDNS
// Client.
package agdc

// UpstreamGroupName is a type for the name of an upstream group.
//
// TODO(e.burkov):  Add validation method, consider using the rules for
// prometheus labels.
type UpstreamGroupName string

const (
	// UpstreamGroupNameDefault is the reserved name for an upstream group that
	// matches all requests and must appear in the configuration.
	UpstreamGroupNameDefault UpstreamGroupName = "default"

	// UpstreamGroupNamePrivate is the reserved name for an upstream group that
	// handles the PTR requests for private IP addresses and must appear in the
	// configuration.
	UpstreamGroupNamePrivate UpstreamGroupName = "private"
)
