// infping.go copyright Tor Hveem
// influxdb2 support by Serafin Rusu
// License: MIT

package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/pelletier/go-toml"
)

func herr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func perr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func slashSplitter(c rune) bool {
	return c == '/'
}

func readPoints(config *toml.Tree, con influxdb2.Client) {
	bindip := config.Get("bindip.srcip").(string)
	args := []string{"-B 1", "-D", "-r0", "-O 0", "-Q 60", "-p 5000", "-l", "-S"}
	args = append(args, bindip)

	hosts := config.Get("hosts.hosts").([]interface{})
	for _, v := range hosts {
		host, _ := v.(string)
		args = append(args, host)
	}
	log.Printf("Going to ping the following hosts: %q", hosts)
	cmd := exec.Command("/usr/sbin/fping", args...)
	stdout, err := cmd.StdoutPipe()
	herr(err)
	stderr, err := cmd.StderrPipe()
	herr(err)
	cmd.Start()
	perr(err)

	buff := bufio.NewScanner(stderr)
	for buff.Scan() {
		text := buff.Text()
		fields := strings.Fields(text)
		// Ignore timestamp
		if len(fields) > 1 {
			host := fields[0]
			data := fields[4]
			dataSplitted := strings.FieldsFunc(data, slashSplitter)
			// Remove ,
			dataSplitted[2] = strings.TrimRight(dataSplitted[2], "%,")
			// lossp := dataSplitted[2]
			sent, recv, lossp := dataSplitted[0], dataSplitted[1], dataSplitted[2]
			min, max, avg := "", "", ""
			// Ping times
			if len(fields) > 5 {
				times := fields[7]
				td := strings.FieldsFunc(times, slashSplitter)
				min, avg, max = td[0], td[1], td[2]
			}
			if config.Get("debug.logs").(bool) {
				log.Printf("Host:%s, loss: %s, min: %s, avg: %s, max: %s", host, lossp, min, avg, max)
			}
			writePoints(config, con, host, sent, recv, lossp, min, avg, max)
		}
	}
	std := bufio.NewReader(stdout)
	line, err := std.ReadString('\n')
	perr(err)
	log.Printf("stdout:%s", line)
}

func writePoints(config *toml.Tree, con influxdb2.Client, host, sent, recv, lossp, min, avg, max string) {
	org := config.Get("influxdb.org").(string)
	bucket := config.Get("influxdb.bucket").(string)
	writeAPI := con.WriteAPIBlocking(org, bucket)

	if min != "" && avg != "" && max != "" {
		minf, _ := strconv.ParseFloat(min, 64)
		avgf, _ := strconv.ParseFloat(avg, 64)
		maxf, _ := strconv.ParseFloat(max, 64)
		lossf, _ := strconv.Atoi(lossp)

		p := influxdb2.NewPointWithMeasurement(config.Get("influxdb.measurement").(string)).
			AddTag("host", host).
			AddField("loss", lossf).
			AddField("min", minf).
			AddField("avg", avgf).
			AddField("max", maxf).
			SetTime(time.Now())
		writeAPI.WritePoint(context.Background(), p)

	} else {
		lossf, _ := strconv.Atoi(lossp)
		p := influxdb2.NewPointWithMeasurement(config.Get("influxdb.measurement").(string)).
			AddTag("host", host).
			AddField("loss", lossf).
			SetTime(time.Now())
		writeAPI.WritePoint(context.Background(), p)
	}

}

func main() {
	config, err := toml.LoadFile("config.toml")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	influx_url := config.Get("influxdb.url").(string)
	token := config.Get("influxdb.token").(string)

	client := influxdb2.NewClient(influx_url, token)

	readPoints(config, client)
}
