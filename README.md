# HeartBeatd

Heatbeatd is a configurable heatbeat that works with [Etcd](https://github.com/coreos/etcd). It could be a companion for [Confd](https://github.com/kelseyhightower/confd).

# Configuration

Create a "`config.yml`" file:

```yaml
etcd: http://127.0.0.1:4001
keys:
    /foo:
        test: http
        interval: 1
        timeout: 10
        value: "{{.Value}}"
        on_fail: "etcdctl rm {{.Key}}"
        on_success: "echo YES"
    /bar:
        test: connect
```

Launching "`heatbeatd`" will read that file, connect to etcd and watch for "`/foo`" and "`/bar`" keys. The wath is recursive.

**Important:** keys should be directory !

As soon as a new key is added, heatbeat creates a "watcher". The watcher use the "test" to test what is recorded in the key value. For example:

```
etcdctd set /foo/A http://www.google.com
```

If the test is ok, "`command_on_success`" is executed. If not, "`command_on_fail`" is executed. 

The given example removes the key - that way, "`confd`" is alerted that a key is dropped and can execute is own templates.

# Where to place config.yml ?

Heartbeatd checks that files in that order:

- `./config.yml`
- `$HOME/.config/heartbeatd/config.yml`
- `/etc/heartbeatd/config.yml`


Heartbeatd will stop if none of these file is found.

# Commands

Be sure that the command is in your "`PATH`"

The command should be a parsable template. The [client.Reponse.Node](https://godoc.org/github.com/coreos/etcd/client#Node) is passed to the template. So you can use values from this struct.


# Yaml values:

```yaml
etcd: "..." # string http url for the etcd server, default "http://127.0.0.1:4001"
interval: 1 # default interval between 2 checks
keys: # map key/test
    /key: #key is the key name to listen
        test: connect|http # at this time, only connect and http
        timeout: 10 # default timeout for http test
        interval: 1 # integer seconds between 2 checks, default 1
        value: "" # template to get value, default {{.Value}}
        on_fail: "..." # command to launch when test fails, default ""
        on_success: "..." # command to laucn when test success, default ""

    /key1: #...
        test: ...
```

You may use several commands in multiline. You must terminate your commands by a coma ";":

```yaml
#...
keys:
    /foo/bar:
        #...
        on_success: |
            command 1 with {{.Key}};
            command 2 with {{.Value}};
```



