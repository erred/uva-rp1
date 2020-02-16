package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
        var file string
        flag.StringVar(&file, "file", "file.conf", "ndn-traffic-client.conf")
        flag.Parse()

        done := make(chan struct{})
        ctx, cancel := context.WithCancel(context.Background())

        cmd := exec.CommandContext(ctx, "ndn-traffic-client", "-i", "5", file)
        out, err := cmd.StdoutPipe()
        if err != nil {
                log.Fatal(err)
        }
        go cmd.Run()
        go func(){
                // global id - time
                m := make(map[int64]int64, 100000)
                // start time - time
                tm := make(map[int64]int64, 100000)
                r := bufio.NewReader(out)
                        cont := true
                for cont {
                        s, err := r.ReadString('\n')
                        if err != nil {
                                cont = false
                                if err != io.EOF {
                                        log.Println(err)
                                }
                        }
                        t, err := time.Parse("2006-Jan-02 15:04:04.999999", s[:27])
                        if err != nil {
                                log.Println("parse time:", s[:27], err)
                                continue
                        }
                        gi := strings.Index(s, "GlobalID=")
                        si := strings.Index(s[gi:], " ")
                        gid, err := strconv.ParseInt(s[gi+9:gi+si], 10, 64)
                        if err != nil {
                                log.Println("parse gid:", s[gi+9:gi+si], err)
                        }
                        if strings.Contains(s, "Sending Interest") {
                                m[gid] = t.UnixNano()
                        } else if strings.Contains(s, "Data Received") {
                                tm[m[gid]] = t.UnixNano() - m[gid]
                        }
                }
                s := make([]int64, 0, len(tm))
                for k := range tm {
                s = append(s, k)
                }
                sort.Slice(s, func(i, j int)bool{
                        return s[i] < s[j]
                })
                for _, k := range s {
                        fmt.Printf("%d,%d\n", k, tm[k])
                }
        }()

        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt)

        <-c
        cancel()
        <-done
}
