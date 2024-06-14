 #  AdGuard DNS Client configuration file

See file [`config.dist.yml`][dist] for a full example of a [YAML][yaml]
configuration file with comments.

##  Contents

 *  [DNS](#dns)
     *  [Cache](#dns-cache)
     *  [Server](#dns-server)
     *  [Bootstrap](#dns-bootstrap)
     *  [Upstream](#dns-upstream)
     *  [Fallback](#dns-fallback)
 *  [Debugging](#debug)
 *  [Logging](#log)

[dist]: ../config.dist.yml
[yaml]: https://yaml.org/



##  <a href="#dns" id="dns" name="dns">DNS</a>

The `dns` object configures the behavior of DNS server.  It has the following
properties:

   ###  <a href="#dns-cache" id="dns-cache" name="dns-cache">Cache</a>

The `cache` object configures caching the results of querying DNS.  It has the
following properties:

 *  <a href="#dns-cache-enabled" id="dns-cache-enabled" name="dns-cache-enabled">`enabled`</a>:
    Whether or not the DNS results should be cached.

    **Example:** `true`.

 *  <a href="#dns-cache-size" id="dns-cache-size" name="dns-cache-size">`size`</a>:
    The maximum size of the DNS results cache as human-readable data size.  It
    must be greater than zero if [`enabled`](#dns-cache-enabled) is `true`.

    **Example:** `128MB`.

 *  <a href="#dns-cache-client-size" id="dns-cache-client-size" name="dns-cache-client-size">`client_size`</a>:
    The maximum size of the DNS results cache for each configured client’s
    address or subnetwork as human-readable data size.  It must be greater than
    zero if [`enabled`](#dns-cache-enabled) is `true`.

    **Example:** `4MB`.



   ###  <a href="#dns-server" id="dns-server" name="dns-server">Server</a>

The `server` object configures the handling of incoming requests.  It has the
following properties:

 *  <a href="#dns-server-listen_addresses" id="dns-server-listen_addresses" name="dns-server-listen_addresses">`listen_addresses`</a>:
    The set of addresses with ports to listen on.

    **Property example:**

    ```yaml
    'listen_addresses':
      - address: '127.0.0.1:53'
      - address: '[::1]:53'
    ```



   ###  <a href="#dns-bootstrap" id="dns-bootstrap" name="dns-bootstrap">Bootstrap</a>

The `bootstrap` object configures the resolving of [upstream](#dns-upstream)
servers addresses.  It has the following properties:

 *  <a href="#dns-bootstrap-servers" id="dns-bootstrap-servers" name="dns-bootstrap-servers">`servers`</a>:
    The list of servers to use for resolving [upstream](#dns-upstream) servers
    hostnames.

    **Property example:**

    ```yaml
    'servers':
      - address: '8.8.8.8:53'
      - address: '192.168.1.1:53'
    ```

 *  <a href="#dns-bootstrap-timeout" id="dns-bootstrap-timeout" name="dns-bootstrap-timeout">`timeout`</a>:
    The timeout for bootstrap DNS requests as a human-readable duration.

    **Example:** `2s`.



   ###  <a href="#dns-upstream" id="dns-upstream" name="dns-upstream">Upstream</a>

The `upstream` object configures the actual resolving of requests.  It has the
following properties:


 *  <a href="#dns-upstream-groups" id="dns-upstream-groups" name="dns-upstream-groups">`groups`</a>:
    The set of upstream servers keyed by the group’s name.  It has the following
    fields:

     *  <a href="#dns-upstream-group-address" id="dns-upstream-group-address" name="dns-upstream-group-address">`address`</a>:
        The upstream server’s address.

        **Example:** `'8.8.8.8:53'`.

     *  <a href="#dns-upstream-group-match" id="dns-upstream-group-match" name="dns-upstream-group-match">`match`</a>:
        The list of criteria to match the request against.  Each entry may
        contain the following properties:

         *  <a href="#dns-upstream-group-match-question_domain" id="dns-upstream-group-match-question_domain" name="dns-upstream-group-match-question_domain">`question_domain`</a>:
            The domain or a suffix of the domain that the set of upstream
            servers should be used to resolve.

            **Example:** `'mycompany.local'`.

         *  <a href="#dns-upstream-group-match-client" id="dns-upstream-group-match-client" name="dns-upstream-group-match-client">`client`</a>:
            The client’s address or a subnet of the client’s address that the
            set of upstream servers should be used to resolve requests from.  It
            must have no significant bits outside of the subnet mask.

            **Example:** `'192.0.2.0/24'`.

        Note, that properties specified within a single entry are combined with
        a logical *AND*.  Entries are combined with a logical *OR*.

        **Property example:**

        ```yaml
        'match':
          - question_domain: 'mycompany.local'
            client: '192.168.1.0/24'
          - question_domain: 'mycompany.external'
          - client: '1.2.3.4'
        ```

    Note that `groups` should contain at least a single entry named `default`,
    and optionally a single entry named `private`, both should have no `match`
    property.

    The `default` group will be used when there are no matches among other
    groups.

    The `private` group will be used to resolve the PTR requests for the private
    IP addresses.  Such queries will be answered with `NXDOMAIN` if no `private`
    group is defined.

 *  <a href="#dns-upstream-timeout" id="dns-upstream-timeout" name="dns-upstream-timeout">`timeout`</a>:
    The timeout for upstream DNS requests as a human-readable duration.

    **Example:** `2s`.



   ###  <a href="#dns-fallback" id="dns-fallback" name="dns-fallback">Fallback</a>

The `fallback` object configures the behavior of DNS server on failures.  It has
the following properties:

 *  <a href="#dns-fallback-servers" id="dns-fallback-servers" name="dns-fallback-servers">`servers`</a>:
    The list of servers to use after the actual [upstream](#dns-upstream) failed
    to respond.

    **Property example:**

    ```yaml
    'servers':
      - address: 'tls://94.140.14.140'
    ```

 *  <a href="#dns-fallback-timeout" id="dns-fallback-timeout" name="dns-fallback-timeout">`timeout`</a>:
    The timeout for fallback DNS requests as a human-readable duration.

    **Example:** `2s`.



##  <a href="#debug" id="debug" name="debug">Debugging</a>

The `debug` object configures the debugging features.  It has the following
properties:

   ###  <a href="#debug-pprof" id="debug-pprof" name="debug-pprof">`pprof`</a>

The `pprof` object configures the [`pprof`][pkg-pprof] HTTP handlers.  It has
the following properties:

 *  <a href="#debug-pprof-port" id="debug-pprof-port" name="debug-pprof-port">`port`</a>:
    The port to listen on for debug HTTP requests on localhost.

    **Example:** `6060`.

 *  <a href="#debug-pprof-enabled" id="debug-pprof-enabled" name="debug-pprof-enabled">`enabled`</a>:
    Whether or not the debug profiling is enabled.

    **Example:** `true`.

[pkg-pprof]: https://golang.org/pkg/net/http/pprof



##  <a href="#log" id="log" name="log">Logging</a>

**NOTE:** In its current state, the log system is only intended for debugging
startup errors.

The `log` object configures the logging.  It has the following properties:

 *  <a href="#log-output" id="log-output" name="log-output">`output`</a>:
    The output to which logs are written.

    **NOTE:** Log entries written to the system log are in text format and use
    the system timestamp.

    Possible values:

     *  `syslog` means that the platform-specific system log is used, which is
        syslog for Linux and Event Log for Windows.

     *  `stdout` for standard output stream.

     *  `stderr` for standard error stream.

     *  Absolute path to the log file.

        **Example:** `/home/user/logs`.

        **Example:** `C:\Users\user\logs.txt`.

    **Example:** `syslog`.

 *  <a href="#log-format" id="log-format" name="log-format">`format`</a>:
    Specifies format of the log entries.

    Possible values:

     *  `adguard_legacy`;
     *  `default`;
     *  `json`;
     *  `jsonhybrid`;
     *  `text`.

    **Example:** `default`.

    <!--
    TODO(s.chzhen):  Add output examples.
    -->

 *  <a href="#log-timestamp" id="log-timestamp" name="log-timestamp">`timestamp`</a>:
    If the log entries should contain timestamp.

    **Example:** `false`.

 *  <a href="#log-verbose" id="log-verbose" name="log-verbose">`verbose`</a>:
    If the log should be more informative.

    **Example:** `false`.
