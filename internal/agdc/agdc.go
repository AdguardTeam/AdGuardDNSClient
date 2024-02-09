// Package agdc contains some types and functions commonly used in AdGuardDNS
// Client.
package agdc

// UpstreamGroupName is a type for the name of an upstream group.
type UpstreamGroupName string

// UpstreamGroupNameDefault is the reserved name for an upstream group that
// matches all requests and must appear in the configuration.
const UpstreamGroupNameDefault UpstreamGroupName = "default"
