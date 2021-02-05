## Simple filesystem perf test plugin

### Building
`go build -o {desired binary name} cmd/main.go`

Place in location accessible to run by telegraf, eg. /usr/local/bin.
eg.

```
go mod init
go build -o telegraf_fsperf cmd/main.go
mv telegraf_dircount /usr/local/bin/
```



### Using
Configure a plugin.conf file with the desired directory to perf test

Make sure `ioping` and `dd` are installed.


eg. `/etc/telegraf_fsperf.conf`
```
[[inputs.fsperf]]
directories = ["/home"]
```

Configure dd.sh, probably in /usr/local/bin. Note the plugin expects the JSON
output created by dd.sh.
```
cp dd.sh /usr/local/bin/
chmod u+x /usr/local/bin/dd.sh
```


Configure as execd plugin in your main telegraf configuration
```
[[inputs.execd]]
command = ["/usr/local/bin/telegraf_fsperf", "-config", "/etc/telegraf_fsperf.conf"]
```

