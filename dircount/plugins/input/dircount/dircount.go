package main

import(
    "os"
)

func main() {
    dir_open, _ := os.Open(`/home`)
    entries, err := dir_open.Readdirnames(-1)
    if err != nil{
        println(entries)
    }
}
