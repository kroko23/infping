# infping

"infping" is a tool that parses the output of fping and stores the results in InfluxDB 2.0.

For more information, please refer to the blog post at https://hveem.no/visualizing-latency-variance-with-grafana

To build the source code (which has been tested on AlmaLinux 8.4), you will need to download and install Go version 1.18 or higher.

Building from source:

Download and install at least v1.18 of go:
```
wget https://go.dev/dl/go1.18.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bash_profile
```
Install git:
```
dnf -y install git
git clone https://github.com/kroko23/infping.git
cd infping
```

Install git modules:
```
go get github.com/influxdata/influxdb-client-go/v2
go get github.com/pelletier/go-toml
```

Build:
```
go build -o infping infping.go
```