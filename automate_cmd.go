package main

import (
    "fmt"
    "log"
    "time"
)

const AutomateReady = "ready"
const AutomateManual = "manual"
const AutomateError = "error"

const AutomateJoinSyncDnskeys = "join-sync-dnskeys"
const AutomateJoinDnskeysSynced = "join-dnskeys-synced"
const AutomateJoinSyncCdscdnskeys = "join-sync-cdscdnskeys"
const AutomateJoinCdscdnskeysSynced = "join-cdscdnskeys-synced"
const AutomateJoinParentDsSynced = "join-parent-ds-synced"
const AutomateJoinRemoveCdscdnskeys = "join-remove-cdscdnskeys"
const AutomateJoinWaitDs = "join-wait-ds"
const AutomateJoinSyncNses = "join-sync-nses"
const AutomateJoinNsesSynced = "join-nses-synced"
const AutomateJoinAddCsync = "join-add-csync"
const AutomateJoinParentNsSynced = "join-parent-ns-synced"
const AutomateJoinRemoveCsync = "join-remove-csync"

const AutomateLeaveSyncNses = "leave-sync-nses"
const AutomateLeaveNsesSynced = "leave-nses-synced"
const AutomateLeaveAddCsync = "leave-add-csync"
const AutomateLeaveParentNsSynced = "leave-parent-ns-synced"
const AutomateLeaveRemoveCsync = "leave-remove-csync"
const AutomateLeaveWaitNs = "leave-wait-ns"
const AutomateLeaveSyncDnskeys = "leave-sync-dnskeys"
const AutomateLeaveDnskeysSynced = "leave-dnskeys-synced"
const AutomateLeaveSyncCdscdnskeys = "leave-sync-cdscdnskeys"
const AutomateLeaveCdscdnskeysSynced = "leave-cdscdnskeys-synced"
const AutomateLeaveParentDsSynced = "leave-parent-ds-synced"
const AutomateLeaveRemoveCdscdnskeys = "leave-remove-cdscdnskeys"

type automation struct {
    Group   string
    Running bool
    Stop    bool
}

var Automation map[string]*automation

func init() {
    Automation = make(map[string]*automation)

    Command["automate-step"] = AutomateStepCmd
    Command["automate-start"] = AutomateStartCmd
    Command["automate-stop"] = AutomateStopCmd
    Command["automate-error"] = AutomateErrorCmd
    Command["automate-clear-error"] = AutomateClearErrorCmd
    Command["automate-autostart"] = AutomateAutostartCmd
    Command["automate-no-autostart"] = AutomateNoAutostartCmd

    CommandHelp["automate-step"] = "Run one step of automation for a group, requires <fqdn>"
    CommandHelp["automate-start"] = "Start automation for a group, requires <fqdn> (only in daemon mode)"
    CommandHelp["automate-stop"] = "Stop automation for a group, requires <fqdn> (only in daemon mode)"
    CommandHelp["automate-error"] = "Show if automation for a group ran into an error, requires <fqdn>"
    CommandHelp["automate-clear-error"] = "Clear the automation error for a group and continue, requires <fqdn> <next step>"
    CommandHelp["automate-autostart"] = "Set automation autostart for a group, requires <fqdn>"
    CommandHelp["automate-no-autostart"] = "Remove automation autostart for a group, requires <fqdn>"
}

func AutomateStepCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    stage := Config.Get("automate-stage:"+args[0], "")

    signers := make(map[string]bool)
    for _, s := range Config.ListGet("signers:" + args[0]) {
        if Config.Get("signer-leaving:"+s, "") == "" {
            signers[s] = false
        } else {
            signers[s] = true
        }
    }
    WsStatus(args[0], stage, signers)

    switch stage {
    case AutomateReady:
        *output = append(*output, "Nothing to do for "+args[0])
        return nil

    case AutomateError:
        *output = append(*output, "Error exist for "+args[0])
        return nil

    case AutomateJoinSyncDnskeys:
        err := SyncDnskeyCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinDnskeysSynced)

    case AutomateJoinDnskeysSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-dnskeys-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "DNSKEYs not synced yet for "+args[0])
            Config.Set("automate-stage:"+args[0], AutomateJoinSyncDnskeys)
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinSyncCdscdnskeys)

    case AutomateJoinSyncCdscdnskeys:
        err := SyncCdscdnskeysCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinCdscdnskeysSynced)

    case AutomateJoinCdscdnskeysSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-cdscdnskeys-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "CDS/CDNSKEYs not synced yet for "+args[0])
            Config.Set("automate-stage:"+args[0], AutomateJoinSyncCdscdnskeys)
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinParentDsSynced)

    case AutomateJoinParentDsSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-parent-ds-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "Parent DS not synced yet for "+args[0])
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinRemoveCdscdnskeys)

    case AutomateJoinRemoveCdscdnskeys:
        err := RemoveCdscdnskeysCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinWaitDs)

    case AutomateJoinWaitDs:
        wait_until := Config.Get("group-wait-ds:"+args[0], "")
        if wait_until == "" {
            err := WaitDsCmd(args, remote, output)
            if err != nil {
                Config.Set("automate-error:"+args[0], err.Error())
                Config.Set("automate-stage:"+args[0], AutomateError)
                return err
            }
        }
        until, err := time.Parse(time.RFC3339, Config.Get("group-wait-ds:"+args[0], ""))
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }

        if time.Now().Before(until) {
            *output = append(*output, fmt.Sprintf("Wait until %s (%s)", until.String(), time.Until(until).String()))
            WsWaitUntil(args[0], time.Until(until).String())
            return nil
        }
        WsWaitUntil(args[0], "done")
        Config.Remove("group-wait-ds:" + args[0])
        Config.Set("automate-stage:"+args[0], AutomateJoinSyncNses)

    case AutomateJoinSyncNses:
        err := SyncNsCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinNsesSynced)

    case AutomateJoinNsesSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-nses-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "NSes not synced yet for "+args[0])
            Config.Set("automate-stage:"+args[0], AutomateJoinSyncNses)
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinAddCsync)

    case AutomateJoinAddCsync:
        err := AddCsyncCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinParentNsSynced)

    case AutomateJoinParentNsSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-parent-ns-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "Parent NS not synced yet for "+args[0])
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateJoinRemoveCsync)

    case AutomateJoinRemoveCsync:
        err := RemoveCsyncCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateReady)

    case AutomateLeaveSyncNses:
        err := SyncNsCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveNsesSynced)

    case AutomateLeaveNsesSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-nses-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "NSes not synced yet for "+args[0])
            Config.Set("automate-stage:"+args[0], AutomateLeaveSyncNses)
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveAddCsync)

    case AutomateLeaveAddCsync:
        err := AddCsyncCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveParentNsSynced)

    case AutomateLeaveParentNsSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-parent-ns-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "Parent NS not synced yet for "+args[0])
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveRemoveCsync)

    case AutomateLeaveRemoveCsync:
        err := RemoveCsyncCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveWaitNs)

    case AutomateLeaveWaitNs:
        wait_until := Config.Get("group-wait-ns:"+args[0], "")
        if wait_until == "" {
            err := WaitNsCmd(args, remote, output)
            if err != nil {
                Config.Set("automate-error:"+args[0], err.Error())
                Config.Set("automate-stage:"+args[0], AutomateError)
                return err
            }
        }
        until, err := time.Parse(time.RFC3339, Config.Get("group-wait-ns:"+args[0], ""))
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }

        if time.Now().Before(until) {
            *output = append(*output, fmt.Sprintf("Wait until %s (%s)", until.String(), time.Until(until).String()))
            WsWaitUntil(args[0], time.Until(until).String())
            return nil
        }
        WsWaitUntil(args[0], "done")
        Config.Remove("group-wait-ns:" + args[0])
        Config.Set("automate-stage:"+args[0], AutomateLeaveSyncDnskeys)

    case AutomateLeaveSyncDnskeys:
        err := SyncDnskeyCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveDnskeysSynced)

    case AutomateLeaveDnskeysSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-dnskeys-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "DNSKEYs not synced yet for "+args[0])
            Config.Set("automate-stage:"+args[0], AutomateLeaveSyncDnskeys)
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveSyncCdscdnskeys)

    case AutomateLeaveSyncCdscdnskeys:
        err := SyncCdscdnskeysCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveCdscdnskeysSynced)

    case AutomateLeaveCdscdnskeysSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-cdscdnskeys-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "CDS/CDNSKEYs not synced yet for "+args[0])
            Config.Set("automate-stage:"+args[0], AutomateLeaveSyncCdscdnskeys)
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveParentDsSynced)

    case AutomateLeaveParentDsSynced:
        err := StatusCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        if synced := Config.Get("group-parent-ds-synced:"+args[0], ""); synced != "yes" {
            *output = append(*output, "Parent DS not synced yet for "+args[0])
            return nil
        }
        Config.Set("automate-stage:"+args[0], AutomateLeaveRemoveCdscdnskeys)

    case AutomateLeaveRemoveCdscdnskeys:
        err := RemoveCdscdnskeysCmd(args, remote, output)
        if err != nil {
            Config.Set("automate-error:"+args[0], err.Error())
            Config.Set("automate-stage:"+args[0], AutomateError)
            return err
        }
        Config.Set("automate-stage:"+args[0], AutomateReady)

    default:
        return fmt.Errorf("Unknown automate stage %s", stage)
    }

    *output = append(*output, "Automate step "+stage+" success, next stage "+Config.Get("automate-stage:"+args[0], "<unknown>"))

    signers = make(map[string]bool)
    for _, s := range Config.ListGet("signers:" + args[0]) {
        if Config.Get("signer-leaving:"+s, "") == "" {
            signers[s] = false
        } else {
            signers[s] = true
        }
    }
    WsStatus(args[0], Config.Get("automate-stage:"+args[0], "<unknown>"), signers)

    return nil
}

func AutomateAutostart() {
    l := Config.ListGet("automate-autostart")
    for _, g := range l {
        output := []string{}
        err := AutomateStartCmd([]string{g}, true, &output)
        if err != nil {
            log.Fatal(err)
        }
        for _, o := range output {
            log.Println("Automate autostart:", o)
        }
    }
}

func AutomateStartCmd(args []string, remote bool, output *[]string) error {
    if !remote {
        return ErrOnlyRemoteCall
    }

    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if _, ok := Automation[args[0]]; ok {
        return fmt.Errorf("Automation for group %s already running", args[0])
    }

    if !Config.ListEntryExists("groups", args[0]) {
        return fmt.Errorf("group %s does not exist", args[0])
    }

    *output = append(*output, "Starting automation for "+args[0])

    a := &automation{
        Group:   args[0],
        Running: true,
        Stop:    false,
    }
    Automation[args[0]] = a

    go func(a *automation) {
        log.Println("Automating", a.Group)
        for !a.Stop {
            time.Sleep(10 * time.Second)

            DaemonLock.Lock()

            args := []string{a.Group}
            output := []string{}
            err := AutomateStepCmd(args, false, &output)
            if err != nil {
                WsConsole("Automation step failed: " + err.Error())
                log.Println("Automation step failed:", err.Error())
                DaemonLock.Unlock()
                continue
            }
            if cerr := Config.Store(DaemonConf); cerr != nil {
                log.Fatal(cerr)
            }
            DaemonLock.Unlock()
            // log.Println("Automation step successful for", a.Group)
            for _, o := range output {
                WsConsole("Automate " + a.Group + ": " + o)
                log.Println("Automate "+a.Group+":", o)
            }
        }
        log.Println("Ending automation for", a.Group)
        a.Running = false
    }(a)

    return nil
}

func AutomateStopCmd(args []string, remote bool, output *[]string) error {
    if !remote {
        return ErrOnlyRemoteCall
    }

    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if _, ok := Automation[args[0]]; !ok {
        return fmt.Errorf("Automation for group %s is not running", args[0])
    }

    *output = append(*output, "Stopping automation for "+args[0])
    Automation[args[0]].Stop = true

    return nil
}

func AutomateErrorCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    stage := Config.Get("automate-stage:"+args[0], "")
    switch stage {
    case "":
        return fmt.Errorf("No automation stage found for %s", args[0])
    case AutomateError:
        error := Config.Get("automate-error:"+args[0], "<unknown>")
        *output = append(*output, "Error during automation for "+args[0]+": "+error)
    default:
        *output = append(*output, "No automation error for "+args[0])
    }

    return nil
}

func AutomateClearErrorCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 2 {
        return fmt.Errorf("requires <fqdn> <next stage>")
    }

    stage := Config.Get("automate-stage:"+args[0], "")
    switch stage {
    case "":
        return fmt.Errorf("No automation stage found for %s", args[0])
    case AutomateError:
    default:
        *output = append(*output, "No automation error for "+args[0])
        return nil
    }

    switch args[1] {
    case AutomateReady:
    case AutomateJoinSyncDnskeys:
    case AutomateJoinDnskeysSynced:
    case AutomateJoinSyncCdscdnskeys:
    case AutomateJoinCdscdnskeysSynced:
    case AutomateJoinParentDsSynced:
    case AutomateJoinRemoveCdscdnskeys:
    case AutomateJoinWaitDs:
    case AutomateJoinSyncNses:
    case AutomateJoinNsesSynced:
    case AutomateJoinAddCsync:
    case AutomateJoinParentNsSynced:
    case AutomateJoinRemoveCsync:
    case AutomateLeaveSyncNses:
    case AutomateLeaveNsesSynced:
    case AutomateLeaveAddCsync:
    case AutomateLeaveParentNsSynced:
    case AutomateLeaveRemoveCsync:
    case AutomateLeaveWaitNs:
    case AutomateLeaveSyncDnskeys:
    case AutomateLeaveDnskeysSynced:
    case AutomateLeaveSyncCdscdnskeys:
    case AutomateLeaveCdscdnskeysSynced:
    case AutomateLeaveParentDsSynced:
    case AutomateLeaveRemoveCdscdnskeys:
    default:
        return fmt.Errorf("Invalid next stage %s", args[1])
    }

    Config.Remove("automate-error:" + args[0])
    Config.Set("automate-stage:"+args[0], args[1])
    *output = append(*output, "Clear automation error for "+args[0]+" and set next stage to "+args[1])

    return nil
}

func AutomateAutostartCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    if !Config.ListEntryExists("groups", args[0]) {
        return fmt.Errorf("group %s does not exist", args[0])
    }

    Config.ListAdd("automate-autostart", args[0], false)
    *output = append(*output, "Autostart enabled for "+args[0])

    return nil
}

func AutomateNoAutostartCmd(args []string, remote bool, output *[]string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires <fqdn>")
    }

    Config.ListRemove("automate-autostart", args[0])
    *output = append(*output, "Autostart disabled for "+args[0])

    return nil
}
