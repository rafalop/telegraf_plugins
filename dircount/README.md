## Simple directory entry count plugin

### Building
`go build -o {desired binary name} cmd/main.go`
eg.
`go build -o telegraf_dircount =cmd/main.go`

Place in location accessible to run by telegraf, eg. /usr/local/bin.


### Using
Configure a plugin.conf file with the desired directory

eg. `/etc/telegraf_dircount.conf`
```
[[inputs.dircount]]
directories = ["/home"]
```


Configure as execd plugin in your main telegraf configuration
```
[[inputs.execd]]
command = ["/usr/local/bin/telegraf_dircount", "-config", "/etc/telegraf_dircount.conf"]
```

