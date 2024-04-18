# AdGuard DNS Client

A cross-platform lightweight DNS client for [AdGuard DNS].  It operates as a DNS
server that forwards DNS requests to the corresponding upstream resolvers.

[AdGuard DNS]: https://adguard-dns.io

## <a href="#start" id="start" name="start">Quick start</a>

Supported platforms:

- macOS;
- Linux;
- Windows.

Supported architectures:

- 64-bit ARM;
- AMD64;
- i386.

### <a href="#start-basic" id="start-basic" name="start-basic">Basic</a>

To run the server as is:

1. Copy the configuration files (only needed once):

    ```sh
    cp ./config.dist.yaml ./config.yaml
    ```

1. Build the `AdGuardDNSClient` binary:

    ```sh
    make go-build
    ```

1. Run it:

    ```sh
    ./AdGuardDNSClient
    ```

## <a href="#opts" id="opts" name="opts">Command-line options</a>

Any option overrides the corresponding value provided by configuration file.

### <a href="#opts-service" id="opts-service" name="opts-service">Service</a>

Option `-s <value>` specifies the OS service action.  Possible values
are:

- `install`: installs AdGuard DNS Client as a service;
- `uninstall`: uninstalls AdGuard DNS Client service;
- `start`: starts the installed AdGuard DNS Client service;
- `stop`: stops the running AdGuard DNS Client;
- `restart`: restarts the running AdGuard DNS Client service.

### <a href="#opts-verbose" id="opts-verbose" name="opts-verbose">Verbose</a>

Option `-v` enables the verbose log output.

## <a href="#conf" id="conf" name="conf">Configuration</a>

### <a href="#conf-file" id="conf-file" name="conf-file">File</a>

The YAML configuration file is described in the [`doc/configuration.md`] file.
Also, there is a sample configuration file `config.dist.yaml`.

[`doc/configuration.md`]: doc/configuration.md

## <a href="#exit-codes" id="exit-codes" name="exit-codes">Exit codes</a>

There are a bunch of different error codes that may appear under different error
conditions:

- `0`: AdGuardDNSClient successfully finished and exited, no errors.

- `1`: Internal error, most probably misconfiguration.

- `2`: Bad command-line argument or its value.

<!-- TODO(e.burkov): Add a few paragraphs about checking the operability. -->

<!-- TODO(e.burkov): Add doc about environment. -->

<!-- TODO(e.burkov): Add GitHub issue templates. -->
