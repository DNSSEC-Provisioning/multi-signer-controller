package main

import (
    "fmt"
)

func init() {
    Command["group-add"] = GroupAddCmd
    Command["group-list"] = GroupListCmd
    Command["group-remove"] = GroupRemoveCmd

    CommandHelp["group-add"] = "Add a new group, requires <fqdn> <parent ip|host> [port]"
    CommandHelp["group-list"] = "List groups"
    CommandHelp["group-remove"] = "Remove a group, can not be in use, requires <fqdn>"
}

func GroupAddCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 2 {
        return fmt.Errorf("requires <fqdn> <parent ip|host> [port]")
    }

    if len(args) == 3 {
        args[1] = args[1] + ":" + args[2]
    } else {
        args[1] = args[1] + ":53"
    }

    if Config.ListAdd("groups", args[0], false) {
        *output = append(*output, fmt.Sprintf("Group %s added", args[0]))

        Config.Set("parent:"+args[0], args[1])
        Config.Set("automate-stage:"+args[0], AutomateReady)
    } else {
        *output = append(*output, fmt.Sprintf("Group %s already exists", args[0]))
    }

    return nil
}

func GroupListCmd(args []string, remote bool, output *[]string) error {
    g := Config.ListGet("groups")

    *output = append(*output, "Groups:")
    for _, v := range g {
        *output = append(*output, fmt.Sprintf("  %s", v))
    }

    return nil
}

func GroupRemoveCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if Config.Exists("signers:" + args[0]) {
        return fmt.Errorf("group %s has signers", args[0])
    }

    if Config.ListRemove("groups", args[0]) {
        *output = append(*output, fmt.Sprintf("Group %s removed", args[0]))
    } else {
        *output = append(*output, fmt.Sprintf("Group %s did not exist", args[0]))
    }

    return nil
}
