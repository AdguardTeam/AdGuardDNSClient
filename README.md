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

2. Build the `AdGuardDNSClient` binary:

    ```sh
    make go-build
    ```

3. Run it:

    ```sh
    ./AdGuardDNSClient
    ```

## <a href="#opts" id="opts" name="opts">Command-line options</a>

Any option overrides the corresponding value provided by configuration file.

### <a href="#opts-verbose" id="opts-verbose" name="opts-verbose">Verbose</a>

   Option <code>-v</code> enables the verbose log output.

## <a href="#conf" id="conf" name="conf">Configuration</a>

### <a href="#conf-file" id="conf-file" name="conf-file">File</a>

The YAML configuration file is described in the [`doc/configuration.md`] file.
Also, there is a sample configuration file `config.dist.yaml`.

[`doc/configuration.md`]: doc/configuration.md

<!-- TODO(e.burkov): Add a few paragraphs about checking the operability. -->

<!-- TODO(e.burkov): Add doc about environment. -->

<!-- TODO(e.burkov): Add GitHub issue templates. -->
