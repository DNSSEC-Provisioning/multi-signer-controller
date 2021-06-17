package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "strings"
    "sync"
)

type config struct {
    m sync.RWMutex

    conf map[string]interface{}

    changed bool
}

var Config = NewConfig()

func NewConfig() *config {
    return &config{
        conf: make(map[string]interface{}),
    }
}

func (c *config) Exists(name string) bool {
    c.m.RLock()
    defer c.m.RUnlock()

    _, ok := c.conf[name]
    return ok
}

func (c *config) Get(name, _default string) string {
    c.m.RLock()
    defer c.m.RUnlock()

    v, ok := c.conf[name]
    if !ok {
        return _default
    }
    if ret, ok := v.(string); ok {
        return ret
    }
    panic(fmt.Sprintf("%s is not string", name))
}

func (c *config) Set(name, value string) {
    c.m.Lock()
    defer c.m.Unlock()

    c.conf[name] = value
    c.changed = true
}

func (c *config) SetIfNotExists(name, value string) bool {
    c.m.Lock()
    defer c.m.Unlock()

    if _, ok := c.conf[name]; !ok {
        c.conf[name] = value
        c.changed = true
        return true
    }
    return false
}

func (c *config) Remove(name string) bool {
    c.m.Lock()
    defer c.m.Unlock()

    _, ok := c.conf[name]
    if !ok {
        return false
    }

    delete(c.conf, name)
    c.changed = true

    return true
}

// Return a list of config keys based on a prefix
func (c *config) PrefixKeys(prefix string) []string {
    c.m.Lock()
    defer c.m.Unlock()

    keys := []string{}
    for k, _ := range c.conf {
        if strings.HasPrefix(k, prefix) {
            keys = append(keys, k)
        }
    }
    return keys
}

func (c *config) ListExists(name string) bool {
    c.m.RLock()
    defer c.m.RUnlock()

    lp, ok := c.conf[name]
    if !ok {
        return false
    }

    _, ok = lp.([]string)
    if !ok {
        panic(fmt.Sprintf("%s is not a list (%T)", name, lp))
    }

    return true
}

func (c *config) ListEntryExists(name, value string) bool {
    c.m.RLock()
    defer c.m.RUnlock()

    lp, ok := c.conf[name]
    if !ok {
        return false
    }

    l, ok := lp.([]string)
    if !ok {
        panic(fmt.Sprintf("%s is not a list (%T)", name, lp))
    }

    for _, v := range l {
        if v == value {
            return true
        }
    }

    return false
}

func (c *config) ListGet(name string) []string {
    c.m.RLock()
    defer c.m.RUnlock()

    lp, ok := c.conf[name]
    if !ok {
        return []string{}
    }

    l, ok := lp.([]string)
    if !ok {
        panic(fmt.Sprintf("%s is not a list (%T)", name, lp))
    }

    return l
}

func (c *config) ListAdd(name, value string, duplicated bool) bool {
    c.m.Lock()
    defer c.m.Unlock()

    lp, ok := c.conf[name]
    if !ok {
        c.conf[name] = []string{value}
        c.changed = true
        return true
    }
    l, ok := lp.([]string)
    if !ok {
        panic(fmt.Sprintf("%s is not a list", name))
    }

    if !duplicated {
        for _, v := range l {
            if v == value {
                return false
            }
        }
    }

    c.conf[name] = append(l, value)
    c.changed = true

    return true
}

func (c *config) ListRemove(name, value string) bool {
    c.m.Lock()
    defer c.m.Unlock()

    lp, ok := c.conf[name]
    if !ok {
        return false
    }
    l, ok := lp.([]string)
    if !ok {
        panic(fmt.Sprintf("%s is not a list", name))
    }

    n := []string{}
    for _, v := range l {
        if v != value {
            n = append(n, v)
        }
    }

    c.conf[name] = n
    c.changed = true

    return true
}

func (c *config) Store(filename string) error {
    c.m.Lock()
    defer c.m.Unlock()

    if !c.changed {
        return nil
    }

    b, err := json.Marshal(c.conf)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(filename, b, 0644)
    if err != nil {
        return err
    }

    c.changed = false

    return nil
}

func (c *config) Load(filename string) error {
    c.m.Lock()
    defer c.m.Unlock()

    b, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }

    conf := make(map[string]interface{})

    err = json.Unmarshal(b, &conf)
    if err != nil {
        return err
    }

    for k, v := range conf {
        switch l := v.(type) {
        case []interface{}:
            n := []string{}
            for _, e := range l {
                s, ok := e.(string)
                if !ok {
                    panic(fmt.Sprintf("conf broken - list %s has entry that is not string (%T)", k, s))
                }
                n = append(n, s)
            }
            conf[k] = n
            break
        default:
            break
        }
    }

    c.conf = conf

    return nil
}
