package main

import (
    "fmt"
)

func init() {
    Command["signer-add"] = SignerAddCmd
    Command["signer-list"] = SignerListCmd
    Command["signer-remove"] = SignerRemoveCmd
    Command["signer-tsig"] = SignerTsigCmd
    Command["signer-mark-leave"] = SignerMarkLeaveCmd
    Command["signer-unmark-leave"] = SignerUnmarkLeaveCmd

    CommandHelp["signer-add"] = "Add a signer to a group, requires <group> <name> <NS fqdn> <ip|host> [port]"
    CommandHelp["signer-list"] = "List signers in a group, requires <group>"
    CommandHelp["signer-remove"] = "Remove signer from a group, requires <group> <name>"
    CommandHelp["signer-tsig"] = "Set or show which TSIG key to use for dynamic updates, requires <name> [TSIG key]"
    CommandHelp["signer-mark-leave"] = "Mark a signer that it's leaving the group"
    CommandHelp["signer-unmark-leave"] = "Unmark a signer that's leaving the group"
}

func SignerAddCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 4 {
        return fmt.Errorf("requires <group> <name> <NS fqdn> <ip|host> [port]")
    }
    if len(args) == 5 {
        args[3] = args[3] + ":" + args[4]
    } else {
        args[3] = args[3] + ":53"
    }

    if !Config.ListEntryExists("groups", args[0]) {
        return fmt.Errorf("group %s does not exist", args[0])
    }

    stage := Config.Get("automate-stage:"+args[0], "")
    if stage != AutomateReady && stage != AutomateManual {
        return fmt.Errorf("group %s is not ready to have more signers (automate stage %s)", args[0], stage)
    }

    if Config.Exists("signer:" + args[1]) {
        return fmt.Errorf("signer %s already exists", args[1])
    }

    Config.Set("signer:"+args[1], args[3])
    Config.Set("signer-ns:"+args[1], args[2])
    Config.Set("signer-group:"+args[1], args[0])
    Config.ListAdd("signers:"+args[0], args[1], false)

    *output = append(*output, fmt.Sprintf("Signer %s added", args[1]))

    if stage != AutomateManual {
        l := Config.ListGet("signers:" + args[0])
        if len(l) > 1 {
            Config.Set("automate-stage:"+args[0], AutomateJoinSyncDnskeys)
            *output = append(*output, fmt.Sprintf("Automation for %s now %s", args[0], AutomateJoinSyncDnskeys))
        }
    }

    return nil
}

func SignerListCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <group>")
    }

    l := Config.ListGet("signers:" + args[0])
    *output = append(*output, fmt.Sprintf("Signers in %s:", args[0]))
    for _, v := range l {
        ip := Config.Get("signer:"+v, "")
        *output = append(*output, fmt.Sprintf("  %s %s", v, ip))
    }

    return nil
}

func SignerRemoveCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <name>")
    }

    if !Config.Exists("signer:" + args[0]) {
        return fmt.Errorf("signer %s does not exists", args[0])
    }

    if Config.Get("signer-leaving:"+args[0], "") != "yes" {
        return fmt.Errorf("signer %s is not marked leaving", args[0])
    }

    group := Config.Get("signer-group:"+args[0], "")

    if !Config.ListEntryExists("groups", group) {
        return fmt.Errorf("group %s does not exist", group)
    }

    if Config.Get("automate-stage:"+group, "") != AutomateReady {
        return fmt.Errorf("group %s is not in ready state", group)
    }

    Config.ListRemove("signers:"+group, args[0])
    Config.Remove("signer:" + args[0])
    Config.Remove("signer-group:" + args[0])
    ns := Config.Get("signer-ns:"+args[0], "")
    Config.Remove("ns-origin:" + ns)
    Config.Remove("signer-ns:" + args[0])
    Config.Remove("signer-tsigkey:" + args[0])
    for _, k := range Config.PrefixKeys("dnskey-origin:") {
        if Config.Get(k, "") == args[0] {
            Config.Remove(k)
        }
    }
    Config.Remove("signer-leaving:" + args[0])

    *output = append(*output, fmt.Sprintf("Signer %s removed", args[0]))

    return nil
}

func SignerTsigCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <name> [TSIG key]")
    }

    if !Config.Exists("signer:" + args[0]) {
        return fmt.Errorf("signer %s does not exist", args[0])
    }

    if len(args) > 1 {
        if !Config.Exists("tsigkey-" + args[1]) {
            return fmt.Errorf("TSIG key does not exist, use conf-set tsigkey-%s <secret>", args[1])
        }

        Config.Set("signer-tsigkey:"+args[0], args[1])
        *output = append(*output, fmt.Sprintf("Signer %s set to use TSIG key %s for dynamic updates", args[0], args[1]))
    } else {
        key := Config.Get("signer-tsigkey:"+args[0], "")
        if key == "" {
            *output = append(*output, fmt.Sprintf("Signer %s has no TSIG key configured", args[0]))
        } else {
            *output = append(*output, fmt.Sprintf("Signer %s is using TSIG key %s for dynamic updates", args[0], key))
        }
    }

    return nil
}

func SignerMarkLeaveCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <name>")
    }

    if !Config.Exists("signer:" + args[0]) {
        return fmt.Errorf("signer %s does not exist", args[0])
    }
    group := Config.Get("signer-group:"+args[0], "")
    if !Config.ListEntryExists("groups", group) {
        return fmt.Errorf("group %s for signer does not exist", group)
    }

    if leaving := Config.Get("signer-leaving:"+args[0], ""); leaving == "yes" {
        *output = append(*output, fmt.Sprintf("Signer %s already leaving", args[0]))
        return nil
    }

    stage := Config.Get("automate-stage:"+group, "")
    if stage != AutomateReady && stage != AutomateManual {
        return fmt.Errorf("group %s is not ready for change (automate stage %s)", group, stage)
    }

    Config.Set("signer-leaving:"+args[0], "yes")
    *output = append(*output, fmt.Sprintf("Signer %s now marked as leaving", args[0]))

    if stage != AutomateManual {
        l := Config.ListGet("signers:" + group)
        if len(l) > 1 {
            Config.Set("automate-stage:"+group, AutomateLeaveSyncNses)
            *output = append(*output, fmt.Sprintf("Automation for %s now %s", group, AutomateLeaveSyncNses))
        }
    }

    return nil
}

func SignerUnmarkLeaveCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <name>")
    }

    if !Config.Exists("signer:" + args[0]) {
        return fmt.Errorf("signer %s does not exist", args[0])
    }
    group := Config.Get("signer-group:"+args[0], "")
    if !Config.ListEntryExists("groups", group) {
        return fmt.Errorf("group %s for signer does not exist", group)
    }

    if leaving := Config.Get("signer-leaving:"+args[0], ""); leaving != "yes" {
        *output = append(*output, fmt.Sprintf("Signer %s is not leaving", args[0]))
        return nil
    }

    stage := Config.Get("automate-stage:"+group, "")
    if stage != AutomateReady && stage != AutomateManual {
        return fmt.Errorf("group %s is not ready for change (automate stage %s)", group, stage)
    }

    Config.Remove("signer-leaving:" + args[0])
    *output = append(*output, fmt.Sprintf("Signer %s is no longer marked as leaving", args[0]))

    if stage != AutomateManual {
        l := Config.ListGet("signers:" + group)
        if len(l) > 1 {
            Config.Set("automate-stage:"+group, AutomateJoinSyncDnskeys)
            *output = append(*output, fmt.Sprintf("Automation for %s now %s", group, AutomateJoinSyncDnskeys))
        }
    }

    return nil
}
