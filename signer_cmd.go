package main

import (
    "fmt"
)

func init() {
    Command["signer-add"] = SignerAddCmd
    Command["signer-list"] = SignerListCmd
    Command["signer-remove"] = SignerRemoveCmd

    CommandHelp["signer-add"] = "Add a signer to a group, requires <group>, <name>, <ip|host> [port]"
    CommandHelp["signer-list"] = "List signers in a group, requires <group>"
    CommandHelp["signer-remove"] = "Remove signer from a group, requires <group>, <name>"
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

    if Config.Exists(fmt.Sprintf("signer:%s", args[1])) {
        return fmt.Errorf("signer %s already exists", args[1])
    }

    Config.Set(fmt.Sprintf("signer:%s", args[1]), args[2])
    Config.ListAdd(fmt.Sprintf("signers:%s", args[0]), args[1], false)

    *output = append(*output, fmt.Sprintf("Signer %s added", args[1]))

    return nil
}

func SignerListCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <group>")
    }

    l := Config.ListGet(fmt.Sprintf("signers:%s", args[0]))
    *output = append(*output, fmt.Sprintf("Signers in %s:", args[0]))
    for _, v := range l {
        ip := Config.Get(fmt.Sprintf("signer:%s", v), "")
        *output = append(*output, fmt.Sprintf("  %s %s", v, ip))
    }

    return nil
}

func SignerRemoveCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 2 {
        return fmt.Errorf("requires <group> <name>")
    }

    if !Config.Exists(fmt.Sprintf("signer:%s", args[1])) {
        return fmt.Errorf("signer %s does not exists", args[1])
    }

    Config.ListRemove(fmt.Sprintf("signers:%s", args[1]), args[0])
    Config.Remove(fmt.Sprintf("signer:%s", args[0]))

    *output = append(*output, fmt.Sprintf("Signer %s removed", args[1]))

    return nil
}
