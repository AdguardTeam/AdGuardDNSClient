 # AdGuard DNS Client

A cross-platform lightweight DNS client for [AdGuard DNS].  It operates as a DNS
server that forwards DNS requests to the corresponding upstream resolvers.

[AdGuard DNS]: https://adguard-dns.io



##  <a href="#start" id="start" name="start">Quick start</a>

Supported platforms:

 *  Darwin-based OSs (macOS);
 *  Linux;
 *  Windows.



   ###  <a href="#start-basic" id="start-basic" name="start-basic">Basic</a>

To run the server as is:

1.  Copy the configuration files (only needed once):

    ```sh
    cp ./config.dist.yaml ./config.yaml
    ```

2.  Build the `AdGuardDNSClient` binary:

    ```sh
    make go-build
    ```

3.  Run it:

    ```sh
    ./AdGuardDNSClient
    ```

<!-- TODO(e.burkov): Add a few paragraphs about checking the operability. -->



##  <a href="#conf" id="conf" name="conf">Configuration</a>

   ###  <a href="#conf-file" id="conf-file" name="conf-file">File</a>

The YAML configuration file is described in the [`doc/configuration.md`] file.

[`doc/configuration.md`]: doc/configuration.md

<!-- TODO(e.burkov): Add sections about environment and . -->
