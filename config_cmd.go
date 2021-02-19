package main

import (
    "fmt"
)

func init() {
    Command["conf-list"] = ConfigListCmd
    CommandHelp["conf-list"] = "List all configured options and their current values"

    Command["conf-set"] = ConfigSetCmd
    CommandHelp["conf-set"] = "Set a config option, requires <name> <value>"
}

func ConfigListCmd(args []string, remote bool, output *[]string) error {
    Config.m.RLock()
    defer Config.m.RUnlock()

    *output = append(*output, "Config:")
    for k, v := range Config.conf {
        *output = append(*output, fmt.Sprintf("  %s: %v", k, v))
    }

    return nil
}

func ConfigSetCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 2 {
        return fmt.Errorf("Missing <name> <value>")
    }

    Config.Set(args[0], args[1])

    *output = append(*output, fmt.Sprintf("Config %s set", args[0]))

    return nil
}
