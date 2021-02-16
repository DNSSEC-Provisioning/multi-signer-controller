package main

import (
    "fmt"
)

func init() {
    Command["signer-add"] = SignerAddCmd
    Command["signer-list"] = SignerListCmd
    Command["signer-remove"] = SignerRemoveCmd
    Command["signer-tsig"] = SignerTsigCmd

    CommandHelp["signer-add"] = "Add a signer to a group, requires <group> <name> <ip|host> [port]"
    CommandHelp["signer-list"] = "List signers in a group, requires <group>"
    CommandHelp["signer-remove"] = "Remove signer from a group, requires <group> <name>"
    CommandHelp["signer-tsig"] = "Set or show which TSIG key to use for dynamic updates, requires <name> [TSIG key]"
}

func SignerAddCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 3 {
        return fmt.Errorf("requires <group> <name> <ip|host> [port]")
    }
    if len(args) == 4 {
        args[2] = args[2] + ":" + args[3]
    } else {
        args[2] = args[2] + ":53"
    }

    if !Config.ListEntryExists("groups", args[0]) {
        return fmt.Errorf("group %s does not exist", args[0])
    }

    if Config.Exists("signer:" + args[1]) {
        return fmt.Errorf("signer %s already exists", args[1])
    }

    Config.Set("signer:"+args[1], args[2])
    Config.ListAdd("signers:"+args[0], args[1], false)

    *output = append(*output, fmt.Sprintf("Signer %s added", args[1]))

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
    if len(args) < 2 {
        return fmt.Errorf("requires <group> <name>")
    }

    if !Config.Exists("signer:" + args[1]) {
        return fmt.Errorf("signer %s does not exists", args[1])
    }

    Config.ListRemove("signers:"+args[1], args[0])
    Config.Remove("signer:" + args[0])

    *output = append(*output, fmt.Sprintf("Signer %s removed", args[1]))

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
