package dircount

import(
    "os"
    "log"
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/inputs"
)

type dirCount struct{
  Directories []string `toml:"directories"`
  ValueName string `toml:"value_name"`
}

func init() {
  inputs.Add("dircount", func() telegraf.Input {
    return &dirCount{
      Directories: []string{},
      ValueName: "entries",
    }
  })
}

func (d *dirCount) Description() string {
  return "Return number of entries in a directory."
}

func (d *dirCount) SampleConfig() string {
  return `
  ## Grabs count of entries in a dir (non recursive)
  [inputs.dircount]
  # List of directories you want to count entries for
  # directories = ["/home", "/var/log"]
  # name for metric
  # value_name = "entries"
`
}

func (d *dirCount) Gather(a telegraf.Accumulator) error {
  d.sendMetric(a)
  return nil
}

func (d *dirCount) getCount(path string) uint64 {
  dir_open, _ := os.Open(path)
  entries, err := dir_open.Readdirnames(-1)
  if err != nil{
    log.Println(err, ` :`,path)
  }
  count := uint64(len(entries))
  return count
}

func (d *dirCount) sendMetric(a telegraf.Accumulator) {
  for _, dir := range d.Directories {
    a.AddFields("dircount", 
                map[string]interface{}{
                  "dir": dir,
                  d.ValueName : d.getCount(dir),
                },
                nil,
    )
  }
}
