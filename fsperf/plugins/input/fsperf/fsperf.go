package fsperf

import(
    "os/exec"
    "os"
    //"strings"
    "log"
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
    "strconv"
    //"time"
    "encoding/json"
)

type fsperf struct{
  Directories []string `toml:"directories"`
  DDscript string `toml:ddscript`
  Timeout string `toml:"timeout"`
  FileSize string `toml:"filesize"`
}

func init() {
  inputs.Add("fsperf", func() telegraf.Input {
    return &fsperf{
      Directories: []string{},
      Timeout: "5",
      FileSize: "500",
      DDscript: "/usr/local/bin/dd.sh",
    }
  })
}

func (d *fsperf) Description() string {
  return "Runs simple filesystem performance test."
}

func (d *fsperf) SampleConfig() string {
  return `
  ## runs short perf test on a directory. must be able
  ## to hold the file size specified for dd test
  [inputs.fsperf]
  # List of directories you want to perf
  # directories = ["/tmp", "/mnt/fastdir"]
  # Max time to spend writing/reading the file
  # timeout = 2
  # Size of file (MiB)
  # filesize = 1024
  # Location of dd script
  # ddscript = /usr/local/bin/dd.sh
`
}

func (d *fsperf) Gather(a telegraf.Accumulator) error {
  d.sendMetric(a)
  return nil
}


type ResultsData struct {
  write uint64
  read uint64
}

func (d *fsperf) getBW(path string) *ResultsData {
  if _, err := os.Stat(d.DDscript); err != nil {
    log.Fatal("Couldn't find required dd script at", d.DDscript)
  }
  var results = new(ResultsData)

  cmd := exec.Command(d.DDscript, path, d.FileSize, d.Timeout)
  out, err := cmd.CombinedOutput()
  if err != nil {
    results.write = 0
    results.read = 0
    return results
  }

  dd_map := make(map[string]json.RawMessage)
  dd_results := make(map[string]interface{})
  err = json.Unmarshal(out, &dd_map)
  err = json.Unmarshal(dd_map["dd"], &dd_results)
  bw_i,_ := strconv.ParseUint(dd_results["write_rate"].(string), 10, 64)
  bw_r_i,_ := strconv.ParseUint(dd_results["read_rate"].(string), 10, 64)

  results.write = bw_i
  results.read = bw_r_i
  return results
}

func (d *fsperf) sendMetric(a telegraf.Accumulator) {
  for _, dir := range d.Directories {
    r := d.getBW(dir)
    a.AddFields("fsperf", 
                map[string]interface{}{
                  "dir": dir,
                  "write": r.write,
                  "read": r.read,
                },
                nil,
    )
  }
}
