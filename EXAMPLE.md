# Run Examples

These examples are all based on `msat1.catch22.se.`, where two PowerDNS
signers (`msg1`, `msg2`) are used and updated using dynamic updates.

At first the group is setup to use a single signer `msg1`, then we add
and join `msg2`, go over each step in the automation for it to completely
join.

Then `msg2` is marked as leaving, go over each step in the automation for it
to completely leave and once done it will be removed from the group.

To simplify, set a short-cut for the base of the command for these examples:

```
CMD="./multi-signer-controller -conf example.conf"
```

## Zone data on setup

On `msg1`:
```
$ORIGIN .
msat1.catch22.se	3600	IN	MX	0 .
msat1.catch22.se	3600	IN	NS	ns1.msg1.catch22.se.
msat1.catch22.se	3600	IN	SOA	msat1.catch22.se hostmaster.msat1.catch22.se 2021061712 300 300 86400 300
msat1.catch22.se	3600	IN	TXT	"v=spf1 -all"
whoami.msat1.catch22.se	3600	IN	TXT	"ns1.msg1.catch22.se Group 1"
```

On `msg2`:
```
$ORIGIN .
msat1.catch22.se	3600	IN	MX	0 .
msat1.catch22.se	3600	IN	NS	ns1.msg2.catch22.se.
msat1.catch22.se	3600	IN	SOA	msat1.catch22.se hostmaster.msat1.catch22.se 101 300 300 86400 300
msat1.catch22.se	3600	IN	TXT	"v=spf1 -all"
whoami.msat1.catch22.se	3600	IN	TXT	"ns1.msg2.catch22.se Group 1"
```

On `parent`:
```
[catch22.se.] msat1.catch22.se. 300 NS ns1.msg1.catch22.se.
[catch22.se.] msat1.catch22.se. 600 DS 38959 13 2 ...
[catch22.se.] msat1.catch22.se. 600 RRSIG DS 13 3 600 20210701090328 20210617073328 36915 catch22.se. ...
[catch22.se.] msat1.catch22.se. 7200 RRSIG NSEC 13 3 7200 20210629065851 20210615052851 36915 catch22.se. ...
[catch22.se.] msat1.catch22.se. 7200 NSEC msg1.catch22.se. NS DS RRSIG NSEC
```

## No daemon, step automation

First we add the group:
```
$CMD group-add msat1.catch22.se. 13.48.238.90
...
2021/06/17 14:48:22 Group msat1.catch22.se. added
```

Next we add the first signer, note that we add the TSIG keys first so that
they exists prior of the signer:
```
$CMD conf-set tsigkey-msat1tsig1 ...tsig-secret...
$CMD conf-set signer-tsigkey:msg1 msat1tsig1
$CMD signer-add msat1.catch22.se. msg1 ns1.msg1.catch22.se. 13.53.206.47
...
2021/06/17 14:49:44 Signer msg1 added
```

Now we add the next signer and this will be the one joining the group,
note how automation stage changed:
```
$CMD conf-set tsigkey-msat1tsig2 ...tsig-secret...
$CMD conf-set signer-tsigkey:msg2 msat1tsig2
$CMD signer-add msat1.catch22.se. msg2 ns1.msg2.catch22.se. 13.53.34.31
...
2021/06/17 14:55:12 Signer msg2 added
2021/06/17 14:55:12 Automation for msat1.catch22.se. now join-sync-dnskeys
```

Now we can run `status` to just check that it can query all server, don't
worry about all the "missing" right now but look at what it found in `msg1`,
`msg2` and the parent:
```
$CMD status msat1.catch22.se.

2021/06/17 14:56:32 Loading config example.conf
2021/06/17 14:56:32 msg1: found DNSKEY 257 3 13 nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 14:56:32 msg1: found DNSKEY 256 3 13 cVsbobWK6RCkVVdmhisDNgor0oRsYyjhv300B9Xdfx7j+WOuAjEEbvt08sAU7u+DSpHE88t8+SGDoefO3DvufQ==
2021/06/17 14:56:32 msg1: found NS ns1.msg1.catch22.se.
2021/06/17 14:56:32 msg2: found DNSKEY 257 3 13 oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 14:56:32 msg2: found DNSKEY 256 3 13 Hmva5+2gzldmznimU5dYnIKYUHAwmROXhrqcxS33eiE2VEnWDJWCrwjjTVxtPzQzcrEXUm4qfx+AqCZR1zVQdw==
2021/06/17 14:56:32 msg2: found NS ns1.msg2.catch22.se.
2021/06/17 14:56:32 Check sync status of msg1 DNSKEYs
2021/06/17 14:56:32 DNSKEY missing in msg2: cVsbobWK6RCkVVdmhisDNgor0oRsYyjhv300B9Xdfx7j+WOuAjEEbvt08sAU7u+DSpHE88t8+SGDoefO3DvufQ==
2021/06/17 14:56:32 Check sync status of msg2 DNSKEYs
2021/06/17 14:56:32 DNSKEY missing in msg1: Hmva5+2gzldmznimU5dYnIKYUHAwmROXhrqcxS33eiE2VEnWDJWCrwjjTVxtPzQzcrEXUm4qfx+AqCZR1zVQdw==
2021/06/17 14:56:32 Check sync status of msg1 CDSes
2021/06/17 14:56:32 CDS missing for KSK: nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 14:56:32 CDS missing for KSK: oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 14:56:32 Check sync status of msg2 CDSes
2021/06/17 14:56:32 CDS missing for KSK: nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 14:56:32 CDS missing for KSK: oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 14:56:32 Check sync status of msg1 CDNSKEYs
2021/06/17 14:56:32 CDNSKEY missing for KSK: nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 14:56:32 CDNSKEY missing for KSK: oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 14:56:32 Check sync status of msg2 CDNSKEYs
2021/06/17 14:56:32 CDNSKEY missing for KSK: nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 14:56:32 CDNSKEY missing for KSK: oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 14:56:32 Check sync status of msg1 NSes
2021/06/17 14:56:32 NS missing: ns1.msg2.catch22.se.
2021/06/17 14:56:32 Check sync status of msg2 NSes
2021/06/17 14:56:32 NS missing: ns1.msg1.catch22.se.
2021/06/17 14:56:32 Check sync status of NSes for leaving signers
2021/06/17 14:56:32 Check sync status of parent 13.48.238.90:53
2021/06/17 14:56:32   found DS 38959 13 2 ce45ee0f2869965a7ec55140d662c7469ea41d50028fbeed6fe1e2f2f92e5ac2
2021/06/17 14:56:32   DS needs removal: 38959 13 2 ce45ee0f2869965a7ec55140d662c7469ea41d50028fbeed6fe1e2f2f92e5ac2
2021/06/17 14:56:32   found NS ns1.msg1.catch22.se.
2021/06/17 14:56:32   Missing NS: ns1.msg2.catch22.se.
```

Next is to start stepping through the automation:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 14:57:23 Syncing msg1 DNSKEYs
2021/06/17 14:57:23 - cVsbobWK6RCkVVdmhisDNgor0oRsYyjhv300B9Xdfx7j+WOuAjEEbvt08sAU7u+DSpHE88t8+SGDoefO3DvufQ==
2021/06/17 14:57:23 nsupdate: Sending inserts 1, removals 0 to signer msg2
2021/06/17 14:57:23 nsupdate: Update took 27.909242ms, rcode NOERROR
2021/06/17 14:57:23   Added DNSKEY to msg2
2021/06/17 14:57:23 Syncing msg2 DNSKEYs
2021/06/17 14:57:23 - Hmva5+2gzldmznimU5dYnIKYUHAwmROXhrqcxS33eiE2VEnWDJWCrwjjTVxtPzQzcrEXUm4qfx+AqCZR1zVQdw==
2021/06/17 14:57:23 nsupdate: Sending inserts 1, removals 0 to signer msg1
2021/06/17 14:57:23 nsupdate: Update took 30.35987ms, rcode NOERROR
2021/06/17 14:57:23   Added DNSKEY to msg1
2021/06/17 14:57:23 Automate step join-sync-dnskeys success, next stage join-dnskeys-synced
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 14:59:59 Automate step join-dnskeys-synced success, next stage join-sync-cdscdnskeys
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:00:16 nsupdate: Sending inserts 4, removals 0 to signer msg1
2021/06/17 15:00:16 nsupdate: Update took 31.181046ms, rcode NOERROR
2021/06/17 15:00:16   Added CDS/CDNSKEYs to msg1
2021/06/17 15:00:16 nsupdate: Sending inserts 4, removals 0 to signer msg2
2021/06/17 15:00:16 nsupdate: Update took 29.614732ms, rcode NOERROR
2021/06/17 15:00:16   Added CDS/CDNSKEYs to msg2
2021/06/17 15:00:16 Automate step join-sync-cdscdnskeys success, next stage join-cdscdnskeys-synced
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:00:26 Automate step join-cdscdnskeys-synced success, next stage join-parent-ds-synced
```

At this stage the parent needs to scan for the CDS/CDNSKEY records and
update it's DS records. This can also be done manually to progress the
automation.
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:00:39 Parent DS not synced yet for msat1.catch22.se.
```

Once the parent's DS records are updated the automation will continue:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:02:30 Automate step join-parent-ds-synced success, next stage join-remove-cdscdnskeys
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:02:46 nsupdate: Sending remove rrset(s) 2 to signer msg1
2021/06/17 15:02:46 nsupdate: Update took 26.615852ms, rcode NOERROR
2021/06/17 15:02:46   Removed CDS/CDNSKEYs from msg1
2021/06/17 15:02:46 nsupdate: Sending remove rrset(s) 2 to signer msg2
2021/06/17 15:02:46 nsupdate: Update took 27.193629ms, rcode NOERROR
2021/06/17 15:02:46   Removed CDS/CDNSKEYs from msg2
2021/06/17 15:02:46 Automate step join-remove-cdscdnskeys success, next stage join-wait-ds
```

Now we need to wait maximum DS TTL * 2:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:02:55 Wait until 2021-06-17 15:22:55 +0200 CEST (19m59.98980194s)
```

This step can be fast-forwarded by changing the until date:
```
$CMD conf-get group-wait-ds:msat1.catch22.se.
...
2021/06/17 15:03:53 Config group-wait-ds:msat1.catch22.se.: 2021-06-17T15:22:55+02:00

$CMD conf-set group-wait-ds:msat1.catch22.se. 2021-06-17T15:02:55+02:00
...
2021/06/17 15:04:09 Config group-wait-ds:msat1.catch22.se. set
```

The waiting time has passed:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:04:33 Automate step join-wait-ds success, next stage join-sync-nses
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:04:44 nsupdate: Sending inserts 2, removals 0 to signer msg1
2021/06/17 15:04:44 nsupdate: Update took 20.837498ms, rcode NOERROR
2021/06/17 15:04:44   Add/rem'ed NSes to msg1
2021/06/17 15:04:44 nsupdate: Sending inserts 2, removals 0 to signer msg2
2021/06/17 15:04:44 nsupdate: Update took 24.628736ms, rcode NOERROR
2021/06/17 15:04:44   Add/rem'ed NSes to msg2
2021/06/17 15:04:44 Automate step join-sync-nses success, next stage join-nses-synced
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:04:54 Automate step join-nses-synced success, next stage join-add-csync
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:05:09 nsupdate: Sending inserts 1, removals 0 to signer msg1
2021/06/17 15:05:09 nsupdate: Update took 22.172967ms, rcode NOERROR
2021/06/17 15:05:09   Added CSYNC to msg1
2021/06/17 15:05:09 nsupdate: Sending inserts 1, removals 0 to signer msg2
2021/06/17 15:05:09 nsupdate: Update took 27.084435ms, rcode NOERROR
2021/06/17 15:05:09   Added CSYNC to msg2
2021/06/17 15:05:09 Automate step join-add-csync success, next stage join-parent-ns-synced
```

At this stage the parent needs to update NS records based on CSYNC. This
can also be done manually to progress the automation.
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:05:18 Parent NS not synced yet for msat1.catch22.se.
```

Once the parent's NS records are updated the automation will continue:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:05:50 Automate step join-parent-ns-synced success, next stage join-remove-csync
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:06:01 nsupdate: Sending remove rrset(s) 1 to signer msg1
2021/06/17 15:06:01 nsupdate: Update took 26.236728ms, rcode NOERROR
2021/06/17 15:06:01   Removed CSYNC from msg1
2021/06/17 15:06:01 nsupdate: Sending remove rrset(s) 1 to signer msg2
2021/06/17 15:06:01 nsupdate: Update took 29.221394ms, rcode NOERROR
2021/06/17 15:06:01   Removed CSYNC from msg2
2021/06/17 15:06:01 Automate step join-remove-csync success, next stage ready
```

At this point the group is in ready state and both signers has joined to
multi-signer group.

If you run the automation again you'll see that there is nothing to do:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:21:09 Nothing to do for msat1.catch22.se.
```

We will now mark `msg2` as a leaving signer and the automation will change:

```
$CMD signer-mark-leave msg2
...
2021/06/17 15:22:17 Signer msg2 now marked as leaving
2021/06/17 15:22:17 Automation for msat1.catch22.se. now leave-sync-nses
```

Let's step through the automation:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:22:49 removing ns1.msg2.catch22.se., leaving signer
2021/06/17 15:22:49 nsupdate: Sending inserts 1, removals 1 to signer msg1
2021/06/17 15:22:49 nsupdate: Update took 29.517875ms, rcode NOERROR
2021/06/17 15:22:49   Add/rem'ed NSes to msg1
2021/06/17 15:22:49 nsupdate: Sending inserts 1, removals 1 to signer msg2
2021/06/17 15:22:49 nsupdate: Update took 28.021756ms, rcode NOERROR
2021/06/17 15:22:49   Add/rem'ed NSes to msg2
2021/06/17 15:22:49 Automate step leave-sync-nses success, next stage leave-nses-synced
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:22:55 Automate step leave-nses-synced success, next stage leave-add-csync
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:23:00 nsupdate: Sending inserts 1, removals 0 to signer msg1
2021/06/17 15:23:00 nsupdate: Update took 25.280774ms, rcode NOERROR
2021/06/17 15:23:00   Added CSYNC to msg1
2021/06/17 15:23:00 nsupdate: Sending inserts 1, removals 0 to signer msg2
2021/06/17 15:23:00 nsupdate: Update took 29.01913ms, rcode NOERROR
2021/06/17 15:23:00   Added CSYNC to msg2
2021/06/17 15:23:00 Automate step leave-add-csync success, next stage leave-parent-ns-synced
```

Once again we need to wait or manually update the parent's NS records:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:23:03 Parent NS not synced yet for msat1.catch22.se.
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:24:06 Automate step leave-parent-ns-synced success, next stage leave-remove-csync
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:24:10 nsupdate: Sending remove rrset(s) 1 to signer msg1
2021/06/17 15:24:10 nsupdate: Update took 28.351098ms, rcode NOERROR
2021/06/17 15:24:10   Removed CSYNC from msg1
2021/06/17 15:24:10 nsupdate: Sending remove rrset(s) 1 to signer msg2
2021/06/17 15:24:10 nsupdate: Update took 23.873898ms, rcode NOERROR
2021/06/17 15:24:10   Removed CSYNC from msg2
2021/06/17 15:24:10 Automate step leave-remove-csync success, next stage leave-wait-ns
```

We wait maximum NS TTL * 2:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:24:12 Wait until 2021-06-17 17:24:12 +0200 CEST (1h59m59.315906269s)
```

Or fast-forward the automation:
```
$CMD conf-get group-wait-ns:msat1.catch22.se.
...
2021/06/17 15:24:23 Config group-wait-ns:msat1.catch22.se.: 2021-06-17T17:24:12+02:00

$CMD conf-set group-wait-ns:msat1.catch22.se. 2021-06-17T15:24:12+02:00
...
2021/06/17 15:24:41 Config group-wait-ns:msat1.catch22.se. set
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:24:44 Automate step leave-wait-ns success, next stage leave-sync-dnskeys
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:24:45 Syncing msg1 DNSKEYs
2021/06/17 15:24:45 - cVsbobWK6RCkVVdmhisDNgor0oRsYyjhv300B9Xdfx7j+WOuAjEEbvt08sAU7u+DSpHE88t8+SGDoefO3DvufQ==
2021/06/17 15:24:45 - Hmva5+2gzldmznimU5dYnIKYUHAwmROXhrqcxS33eiE2VEnWDJWCrwjjTVxtPzQzcrEXUm4qfx+AqCZR1zVQdw==
2021/06/17 15:24:45 Signer msg2 is leaving, removing it's DNSKEYs from others
2021/06/17 15:24:45 - Hmva5+2gzldmznimU5dYnIKYUHAwmROXhrqcxS33eiE2VEnWDJWCrwjjTVxtPzQzcrEXUm4qfx+AqCZR1zVQdw==
2021/06/17 15:24:45 nsupdate: Sending inserts 0, removals 1 to signer msg1
2021/06/17 15:24:45 nsupdate: Update took 20.755376ms, rcode NOERROR
2021/06/17 15:24:45   Removed DNSKEY from msg1
2021/06/17 15:24:45 Automate step leave-sync-dnskeys success, next stage leave-dnskeys-synced
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:24:48 Automate step leave-dnskeys-synced success, next stage leave-sync-cdscdnskeys
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:24:51 nsupdate: Sending inserts 2, removals 0 to signer msg1
2021/06/17 15:24:51 nsupdate: Update took 29.548505ms, rcode NOERROR
2021/06/17 15:24:51   Added CD/CDNSKEYs to msg1
2021/06/17 15:24:51 Automate step leave-sync-cdscdnskeys success, next stage leave-cdscdnskeys-synced
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:25:01 Automate step leave-cdscdnskeys-synced success, next stage leave-parent-ds-synced
```

And now we need to update the parent's DS records:
```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:25:06 Parent DS not synced yet for msat1.catch22.se.
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:25:24 Automate step leave-parent-ds-synced success, next stage leave-remove-cdscdnskeys
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:25:25 nsupdate: Sending remove rrset(s) 2 to signer msg1
2021/06/17 15:25:25 nsupdate: Update took 21.015897ms, rcode NOERROR
2021/06/17 15:25:25   Removed CD/CDNSKEYs from msg1
2021/06/17 15:25:25 nsupdate: Sending remove rrset(s) 2 to signer msg2
2021/06/17 15:25:25 nsupdate: Update took 16.313781ms, rcode NOERROR
2021/06/17 15:25:25   Removed CD/CDNSKEYs from msg2
2021/06/17 15:25:25 Automate step leave-remove-cdscdnskeys success, next stage ready
```

```
$CMD automate-step msat1.catch22.se.
...
2021/06/17 15:25:26 Nothing to do for msat1.catch22.se.
```

Finally `msg2` has been phased out from the group and can be removed:
```
$CMD signer-remove msg2
...
2021/06/17 15:25:54 Signer msg2 removed
```

## Daemon, full automation

For full automation we will run the daemon and send all commands over an
gRPC, so we need to update the short-cut:

```
CMD="./multi-signer-controller -conf example2.conf -remote 127.0.0.1:5353"
```
Note the usage of a new configuration file.

Now we start the daemon with the HTTP status service (use an IP/port you can connect to):
```
./multi-signer-controller -conf example2.conf -http 127.0.0.1:8080 daemon 127.0.0.1:5353
```

Open a browser to the HTTP address so you can view the console.

Then we add the group:
```
$CMD group-add msat1.catch22.se. 13.48.238.90
...
2021/06/17 15:38:34 Group msat1.catch22.se. added
```

Next we add the first signer, note that we add the TSIG keys first so that
they exists prior of the signer:
```
$CMD conf-set tsigkey-msat1tsig1 ...tsig-secret...
$CMD conf-set signer-tsigkey:msg1 msat1tsig1
$CMD signer-add msat1.catch22.se. msg1 ns1.msg1.catch22.se. 13.53.206.47
...
2021/06/17 15:39:30 Signer msg1 added
```

Now we add the next signer and this will be the one joining the group,
note how automation stage changed:
```
$CMD conf-set tsigkey-msat1tsig2 ...tsig-secret...
$CMD conf-set signer-tsigkey:msg2 msat1tsig2
$CMD signer-add msat1.catch22.se. msg2 ns1.msg2.catch22.se. 13.53.34.31
...
2021/06/17 15:43:09 Signer msg2 added
2021/06/17 15:43:09 Automation for msat1.catch22.se. now join-sync-dnskeys
```

Now we can run `status` to just check that it can query all server, don't
worry about all the "missing" right now but look at what it found in `msg1`,
`msg2` and the parent:
```
$CMD status msat1.catch22.se.
...
2021/06/17 15:43:36 msg1: found DNSKEY 257 3 13 nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 15:43:36 msg1: found DNSKEY 256 3 13 cVsbobWK6RCkVVdmhisDNgor0oRsYyjhv300B9Xdfx7j+WOuAjEEbvt08sAU7u+DSpHE88t8+SGDoefO3DvufQ==
2021/06/17 15:43:36 msg1: found NS ns1.msg1.catch22.se.
2021/06/17 15:43:36 msg2: found DNSKEY 257 3 13 oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 15:43:36 msg2: found DNSKEY 256 3 13 Hmva5+2gzldmznimU5dYnIKYUHAwmROXhrqcxS33eiE2VEnWDJWCrwjjTVxtPzQzcrEXUm4qfx+AqCZR1zVQdw==
2021/06/17 15:43:36 msg2: found NS ns1.msg2.catch22.se.
2021/06/17 15:43:36 Check sync status of msg1 DNSKEYs
2021/06/17 15:43:36 DNSKEY missing in msg2: cVsbobWK6RCkVVdmhisDNgor0oRsYyjhv300B9Xdfx7j+WOuAjEEbvt08sAU7u+DSpHE88t8+SGDoefO3DvufQ==
2021/06/17 15:43:36 Check sync status of msg2 DNSKEYs
2021/06/17 15:43:36 DNSKEY missing in msg1: Hmva5+2gzldmznimU5dYnIKYUHAwmROXhrqcxS33eiE2VEnWDJWCrwjjTVxtPzQzcrEXUm4qfx+AqCZR1zVQdw==
2021/06/17 15:43:36 Check sync status of msg1 CDSes
2021/06/17 15:43:36 CDS missing for KSK: nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 15:43:36 CDS missing for KSK: oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 15:43:36 Check sync status of msg2 CDSes
2021/06/17 15:43:36 CDS missing for KSK: nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 15:43:36 CDS missing for KSK: oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 15:43:36 Check sync status of msg1 CDNSKEYs
2021/06/17 15:43:36 CDNSKEY missing for KSK: nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 15:43:36 CDNSKEY missing for KSK: oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 15:43:36 Check sync status of msg2 CDNSKEYs
2021/06/17 15:43:36 CDNSKEY missing for KSK: nNKWCP5WcfkgC391ZCaUs5yYggwiB21U+teHyJZj2c+8lrwSfp5ENr99EwqkMp4kfHWRb7/M6sVevp9yISyPlA==
2021/06/17 15:43:36 CDNSKEY missing for KSK: oVlyvr3PcPsLxLnMYcsUrvOQ+fQOoqgT927RUB4Sk0Sc7MG3D14/QBvA3k7+I1G2ho2oUU5LIkt1PZmaOZAOkQ==
2021/06/17 15:43:36 Check sync status of msg1 NSes
2021/06/17 15:43:36 NS missing: ns1.msg2.catch22.se.
2021/06/17 15:43:36 Check sync status of msg2 NSes
2021/06/17 15:43:36 NS missing: ns1.msg1.catch22.se.
2021/06/17 15:43:36 Check sync status of NSes for leaving signers
2021/06/17 15:43:36 Check sync status of parent 13.48.238.90:53
2021/06/17 15:43:36   found DS 38959 13 2 ce45ee0f2869965a7ec55140d662c7469ea41d50028fbeed6fe1e2f2f92e5ac2
2021/06/17 15:43:36   DS needs removal: 38959 13 2 ce45ee0f2869965a7ec55140d662c7469ea41d50028fbeed6fe1e2f2f92e5ac2
2021/06/17 15:43:36   found NS ns1.msg1.catch22.se.
2021/06/17 15:43:36   Missing NS: ns1.msg2.catch22.se.
```

If you reload the dashboard you will also see that the group has appeared.

To start the automation run the following:
```
$CMD automate-start msat1.catch22.se.
...
2021/06/17 15:45:55 Starting automation for msat1.catch22.se.
```
You can also enable autostart of automation, see command `automate-autostart`.

The automation will now progress for the group, you can view the progress in
the console.

Once the automation hits `join-parent-ds-synced`, `join-parent-ns-synced`,
`leave-parent-ns-synced` or `leave-parent-ds-synced` it will wait for the
parent's DS/NS records to be updated based on the signers CDS/CDNSKEY/CSYNC.
If there is no automation at the parent you'll need to do this manually to
progress.

There are also two waiting stages `join-wait-ds` and `leave-wait-ns`, these
can be fast-forwarded by setting the wait time to the past. See the stepping
automation example above for how that is done.

Once the group is `ready` you can mark `msg2` as leaving signer and follow
the automation for when a signers is to be removed.

```
$CMD signer-mark-leave msg2
...
2021/06/17 15:55:53 Signer msg2 now marked as leaving
2021/06/17 15:55:53 Automation for msat1.catch22.se. now leave-sync-nses
```

Then once again when the group is `ready` the `msg2` signer can be removed:
```
$CMD signer-remove msg2
...
2021/06/17 15:58:59 Signer msg2 removed
```
