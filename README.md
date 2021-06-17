# multi-signer-controller
Control of a DNSSEC multi-signer group

# Configuration

All configuration option values are either a `string` or an array of `string`.

Current list of configuration options:
- `groups`: An array of all multi-signer groups as FQDNs
- `signers:<fqdn>`: An array of all signer names in a group.
- `signer:<name>`: The `<host|ip>:port` of the authority name-server of a signer.
- `signer-group:<name>`: The FQDN of the group a signer is part of.
- `signer-type:<name>`: The type of Updater to use for the signer, default `nsupdate`.
- `signer-ns:<name>`: The FQDN of the NS for a signer.
- `signer-tsigkey:<name>`: The name of the TSIG key to use.
- `signer-desec:<name>`: The name of the deSEC.io token to use.
- `signer-leaving:<name>`: Exists if the signer is leaving the group.
- `parent:<fqdn>`: The `<host|ip>:port` of the parent of a group.
- `group-ttl:<fqdn>`: The TTL to use when creating new resource records for a group.
- `group-dnskeys-synced:<fqdn>`: Exists if the DNSKEYs are synced within a group.
- `group-cdscdnskeys-synced:<fqdn>`: Exists if the CDS/CDNSKEYs are synced within a group.
- `group-nses-synced:<fqdn>`: Exists if the NSes are synced within a group.
- `group-parent-ds-synced:<fqdn>`: Exists if the parent's DS is in sync with the group.
- `group-parent-ns-synced:<fqdn>`: Exists if the parent's NS is in sync with the group.
- `group-wait-ds:<fqdn>`: An RFC3399 date that exists if the group is waiting for DS records to propagate.
- `group-wait-ns:<fqdn>`: An RFC3399 date that exists if the group is waiting for NS records to propagate.
- `automate-stage:<fqdn>`: The current stage of the automation.
- `automate-error:<fqdn>`: Exists if the automation ran into an error, if so it contains the string of an `error`.
- `dnskey-origin:<dnskey>`: Set during sync when new DNSKEYs are detected, will contain the signer it was seen in.
- `ns-origin:<ns fqdn>`: Set during sync when new NSes are detected, will contain the signer it was seen in.
- `tsigkey-<name>`: The secret of a TSIG key.
- `desectoken-<name>`: The secret of a deSEC.io token.
- `debug-updater`: Set to `yes` to enable debug output of updaters.

# Updaters

To update the signers there are different kinds of updaters, using the
`Updater` interface.

Available updaters:
- `nsupdate`: Uses dynamic updates to change zone information, requires a valid TSIG key to be configured.
- `desec`: Uses deSEC.io API, not implemented yet.

# Automation Stages

Following automation stages exists.

- `ready`: The group is ready for to receive changes.
- `manual`: Can be used to mark a group as manually changed.
- `error`: The automation encountered an error, see command `automate-error`.

- `join-sync-dnskeys`: A new signer has joined and the DNSKEYs needs to be synced.
- `join-dnskeys-synced`: Check that the DNSKEYs are in sync.
- `join-sync-cdscdnskeys`: The CDS/CDNSKEYs needs to be created/synced.
- `join-cdscdnskeys-synced`: Check that the CDS/CDNSKEYs are in sync.
- `join-parent-ds-synced`: Check that the parent's DS are in sync.
- `join-remove-cdscdnskeys`: Remove CDS/CDNSKEYs.
- `join-wait-ds`: Wait for DS to propagate.
- `join-sync-nses`: The NSes needs to be created/synced.
- `join-nses-synced`: Check that the NSes are in sync.
- `join-add-csync`: Add CSYNC.
- `join-parent-ns-synced`: Check that the parent's NS are in sync.
- `join-remove-csync`: Remove CSYNC.

- `leave-sync-nses`: A signer is leaving and the NSes needs to be synced.
- `leave-nses-synced`: Check that the NSes are in sync.
- `leave-add-csync`: Add CSYNC.
- `leave-parent-ns-synced`: Check that the parent's NS are in sync.
- `leave-remove-csync`: Remove CSYNC.
- `leave-wait-ns`: Wait for NS to propagate.
- `leave-sync-dnskeys`: The DNSKEYs needs to be removed/synced.
- `leave-dnskeys-synced`: Check that the DNSKEYs are in sync.
- `leave-sync-cdscdnskeys`: The CDS/CDNSKEYs needs to be created/synced.
- `leave-cdscdnskeys-synced`: Check that the CDS/CDNSKEYs are in sync.
- `leave-parent-ds-synced`: Check that the parent's DS are in sync.
- `leave-remove-cdscdnskeys`: Remove CDS/CDNSKEYs.


# Runtime

*multi-signer-controller* requires `-conf` to be specified at runtime, you can
see all runtime options by using `-help`.

## Running a daemon

*multi-signer-controller* can be run as a daemon with `daemon` command, once
started another *multi-signer-controller* can communicate with that daemon
by using `-remote`.

When running as daemon you can also enable the web-based status interface
by specifying a HTTP listening address and port using `-http`.

# Commands

All commands help and required parameters can be view using the `help`
command (without dash).

# Example: How to run

See `EXAMPLE.md`.

# local go

```
mkdir -p go/1.16.5; wget -O - https://storage.googleapis.com/golang/go1.16.5.linux-amd64.tar.gz | tar -C go/1.16.5 -zxv
export GOROOT="$HOME/go/1.16.5/go" GOPATH="$HOME/go"
export PATH="$PATH:$GOROOT/bin"
```

# build

```
make
./multi-signer-controller
```

# Known Issues

- TSIG keys hardcoded to HMAC-SHA256
- deSEC.io support not implemented yet
