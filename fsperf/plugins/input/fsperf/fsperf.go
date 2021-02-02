package fsperf

import(
    "os/exec"
    "os"
    "strings"
    "log"
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
    "strconv"
    "time"
    "encoding/json"
)

type fsperf struct{
  Directories []string `toml:"directories"`
  DDscript string `toml:ddscript`
  Timeout string `toml:"timeout"`
  FileSize string `toml:"filesize"`
  RandIoTime string `toml:"randio_time"`
  RunInterval int `toml:"run_interval"`
}

func init() {
  inputs.Add("fsperf", func() telegraf.Input {
    return &fsperf{
      Directories: []string{},
      Timeout: "20",
      FileSize: "500",
      DDscript: "/usr/local/bin/dd.sh",
      RandIoTime: "5",
      RunInterval: 10,
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
  # filesize = 500
  # Location of dd script
  # ddscript = /usr/local/bin/dd.sh
  # Time to spend doing randomio
  # randio_time = 5
  # Interval between runs in minutes
  # run_interval = 10
`
}

func (d *fsperf) Gather(a telegraf.Accumulator) error {
  d.sendMetric(a)
  time.Sleep(time.Duration(d.RunInterval)*time.Minute)
  return nil
}


type ResultsData struct {
  write uint64
  read uint64
  rand_iops_cache uint64
  rand_iops_nocache uint64
  rand_nocache_maxlat uint64
  rand_nocache_avglat uint64
  rand_cache_avglat uint64
  rand_cache_maxlat uint64
}

func (d *fsperf) getBW(path string, results *ResultsData) {
  if _, err := os.Stat(d.DDscript); err != nil {
    log.Fatal("Couldn't find required dd script at", d.DDscript)
  }

  cmd := exec.Command(d.DDscript, path, d.FileSize, d.Timeout)
  out, err := cmd.CombinedOutput()
  if err != nil {
    results.write = 0
    results.read = 0
    return
  }

  dd_map := make(map[string]json.RawMessage)
  dd_results := make(map[string]interface{})
  err = json.Unmarshal(out, &dd_map)
  err = json.Unmarshal(dd_map["dd"], &dd_results)
  bw_i,_ := strconv.ParseUint(dd_results["write_rate"].(string), 10, 64)
  bw_r_i,_ := strconv.ParseUint(dd_results["read_rate"].(string), 10, 64)

  results.write = bw_i
  results.read = bw_r_i
  return
}

//Run ioping to test random io perf
func (d *fsperf) getSmallIO(path string, results *ResultsData) {
  ioping_bin, err := exec.LookPath("ioping")
  if err != nil {
    log.Fatal("Error finding ioping. Is it installed?")
  }

  // Run ioping with nocache
  cmd := exec.Command(ioping_bin, `-G`, `-i0`, `-B`, `-s4k`, `-w`+d.RandIoTime, path)
  out, err := cmd.CombinedOutput()
  if err != nil {
    results.rand_iops_nocache = 0
  }
  ioping_results := strings.Split(string(out), " ") 
  results.rand_iops_nocache, _ = strconv.ParseUint(ioping_results[2], 10, 64)
  results.rand_nocache_avglat, _ = strconv.ParseUint(ioping_results[5], 10, 64)
  results.rand_nocache_maxlat, _ = strconv.ParseUint(ioping_results[6], 10, 64)

  // Run ioping with cache
  cmd = exec.Command(ioping_bin, `-C`, `-G`, `-i0`, `-B`, `-s4k`, `-w`+d.RandIoTime, path)
  out, err = cmd.CombinedOutput()
  if err != nil {
    results.rand_iops_cache = 0
  }
  ioping_results = strings.Split(string(out), " ") 
  results.rand_iops_cache, _ = strconv.ParseUint(ioping_results[2], 10, 64)
  results.rand_cache_avglat, _ = strconv.ParseUint(ioping_results[5], 10, 64)
  results.rand_cache_maxlat, _ = strconv.ParseUint(ioping_results[6], 10, 64)
  return
}

func (d *fsperf) fetchResults(dir string, results_chan chan map[string]interface{}) {
  var results = new(ResultsData)
  d.getBW(dir, results)
  d.getSmallIO(dir, results)
  results_chan <- map[string]interface{}{
    "write": results.write,
    "read": results.read,
    "rand_cache": results.rand_iops_cache,
    "rand_nocache": results.rand_iops_nocache,
    "rand_nocache_maxlat": results.rand_nocache_maxlat,
    "rand_nocache_avglat": results.rand_nocache_avglat,
    "rand_cache_maxlat": results.rand_cache_maxlat,
    "rand_cache_avglat": results.rand_cache_avglat,
  }
  return
}

func (d *fsperf) sendMetric(a telegraf.Accumulator) {
  var results_chan = make(chan map[string]interface{})
  for _, dir := range d.Directories {
    go d.fetchResults(dir, results_chan)
  }
  for _,dir := range d.Directories {
    a.AddFields("fsperf", <-results_chan, map[string]string{"dir": dir})
  }
}
