package main

import (
    "encoding/json"
    "io/ioutil"
    "sync"
)

type config struct {
    m sync.RWMutex

    conf map[string]string

    changed bool
}

var Config = NewConfig()

func NewConfig() *config {
    return &config{
        conf: make(map[string]string),
    }
}

func (c *config) Get(name, _default string) string {
    c.m.RLock()
    defer c.m.RUnlock()

    v, ok := c.conf[name]
    if !ok {
        return _default
    }
    return v
}

func (c *config) Set(name, value string) {
    c.m.Lock()
    defer c.m.Unlock()

    c.conf[name] = value

    c.changed = true
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

    conf := make(map[string]string)

    err = json.Unmarshal(b, &conf)
    if err != nil {
        return err
    }

    c.conf = conf

    return nil
}
