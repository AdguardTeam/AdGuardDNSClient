# AdGuard DNS Client

<div align="center">
    <picture>
        <source media="(prefers-color-scheme: dark)" srcset="https://cdn.adtidy.org/website/images/AdGuardDNS_light.svg">
        <img alt="AdGuard DNS Logo" src="https://cdn.adtidy.org/website/images/AdGuardDNS_black.svg" width="300px"/>
    </picture>
</div>

<p align="center">
    <img alt="Screenshot showing the logs of AdGuard DNS Client" src="https://cdn.adtidy.org/content/illustrations/adguard_dns_client.png" width="800"/>
</p>

A cross-platform lightweight DNS client for [AdGuard DNS]. It operates as a DNS server that forwards DNS requests to the corresponding upstream resolvers.

[AdGuard DNS]: https://adguard-dns.io

## <a href="#start" id="start" name="start">Quick start</a>

> [!NOTE]
> AdGuard DNS Client is still in the Beta stage. Things will break and there are still bugs.

Supported operating systems:

- Linux;
- macOS;
- Windows.

Supported CPU architectures:

- 64-bit ARM;
- AMD64;
- i386.

## <a href="#start-basic" id="start-basic" name="start-basic">Getting started</a>

### <a href="#start-basic-unix" id="start-basic-unix" name="start-basic-unix">Unix-like</a>

1. Download and unpack the `.tar.gz` or `.zip` archive from the [releases page][releases].

2. Install it as a service by running:

    ```sh
    ./AdGuardDNSClient -s install -v
    ```

3. Edit the configuration file `config.yaml`.

4. Start the service:

    ```sh
    ./AdGuardDNSClient -s start -v
    ```

To check that it works, use any DNS checking utility. For example, using `nslookup`:

```sh
nslookup -debug 'www.example.com' '127.0.0.1'
```

[releases]: https://github.com/AdguardTeam/AdGuardDNSClient/releases

### <a href="#start-basic-win" id="start-basic-win" name="start-basic-win">Windows</a>

Just download and install using the MSI installer from the [releases page][releases].

To check that it works, use any DNS checking utility. For example, using `nslookup.exe`:

```sh
nslookup -debug "www.example.com" "127.0.0.1"
```

## <a href="#dev" id="dev" name="dev">Developing and contributing</a>

See [`CONTRIBUTING.md`][contr] for more details on how to contribute.

[contr]: ./CONTRIBUTING.md

### <a href="#dev-start" id="dev-start" name="dev-start">Development quick start</a>

You will need Go 1.22 or later. First, register our pre-commit hooks:

```sh
make init
```

Then, install the necessary tools and dependencies:

```sh
make go-deps go-tools
```

That’s pretty much it! You should now be able to lint, test, and build the `AdGuardDNSClient` binary:

```sh
make go-lint
make go-test
make go-build
```

For building packages, you might need additional tools, such as GnuPG, MSI Tools (v0.103 and later), etc.  See `./scripts/make/build-release.sh`.

## <a href="#opts" id="opts" name="opts">Command-line options</a>

Any option overrides the corresponding value provided by configuration file.

### <a href="#opts-service" id="opts-service" name="opts-service">Service</a>

Option `-s <value>` specifies the OS service action. Possible values are:

- `install`: installs AdGuard DNS Client as a service;
- `restart`: restarts the running AdGuard DNS Client service.
- `start`: starts the installed AdGuard DNS Client service;
- `status`: shows the status of the installed AdGuard DNS Client service;
- `stop`: stops the running AdGuard DNS Client;
- `uninstall`: uninstalls AdGuard DNS Client service;

### <a href="#opts-verbose" id="opts-verbose" name="opts-verbose">Verbose</a>

Option `-v` enables the verbose log output.

## <a href="#conf" id="conf" name="conf">Configuration</a>

### <a href="#conf-file" id="conf-file" name="conf-file">File</a>

The YAML configuration file is described in the [`doc/configuration.md`] file,
and there is also a sample configuration file `config.dist.yaml`.  Some
configuration parameters can also be overridden using the environment, see
[`doc/environment.md`].

[`doc/configuration.md`]: doc/configuration.md
[`doc/environment.md`]:   doc/environment.md

## <a href="#exit-codes" id="exit-codes" name="exit-codes">Exit codes</a>

There are a bunch of different error codes that may appear under different error
conditions:

- `0`: AdGuardDNSClient successfully finished and exited, no errors.

- `1`: Internal error, most probably misconfiguration.

- `2`: Bad command-line argument or its value.
