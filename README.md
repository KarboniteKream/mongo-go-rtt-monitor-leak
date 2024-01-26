# mongo-go-rtt-monitor-leak

JIRA: https://jira.mongodb.org/browse/GODRIVER-3107

This repository reproduces a connection leak in `rtt_monitor.go` with
`mongo-go-client` v1.13.1 (and probably also v1.13.0). The issue is not
reproducible on v1.12.2.

This application connects to a Mongo cluster, and creates a HTTP server
listening on port 19508. Every time http://localhost:19508/ping is loaded, it
will disconnect from Mongo and establish a new connection.

First, update `URI` in `main.go`, and run with:
```bash
$ go run main.go pinger.go
```

As soon as the application is started, we can confirm that a Goroutine with
`runHellos` exists for each node in the cluster):
```bash
$ curl -s http://localhost:19508/debug/pprof/goroutine?debug=2 | grep "runHellos"
go.mongodb.org/mongo-driver/x/mongo/driver/topology.(*rttMonitor).runHellos(0x14000098140, 0x14000202c80)
go.mongodb.org/mongo-driver/x/mongo/driver/topology.(*rttMonitor).runHellos(0x14000098280, 0x140003d6280)
go.mongodb.org/mongo-driver/x/mongo/driver/topology.(*rttMonitor).runHellos(0x140000981e0, 0x140002e8000)
```

However, after one or more times the connection is disconnected, we can observe
that the number of `runHellos` Goroutines slowly keeps increasing:
```bash
$ curl -s http://localhost:19508/ping
$ curl -s http://localhost:19508/debug/pprof/goroutine?debug=2 | grep "runHellos"
go.mongodb.org/mongo-driver/x/mongo/driver/topology.(*rttMonitor).runHellos(0x1400023d0e0, 0x1400031b180)
go.mongodb.org/mongo-driver/x/mongo/driver/topology.(*rttMonitor).runHellos(0x140003b65a0, 0x14000548000)
go.mongodb.org/mongo-driver/x/mongo/driver/topology.(*rttMonitor).runHellos(0x140003b66e0, 0x14000263180)
go.mongodb.org/mongo-driver/x/mongo/driver/topology.(*rttMonitor).runHellos(0x140003b6640, 0x14000486c80)
```

All of these are waiting here:
https://github.com/mongodb/mongo-go-driver/blob/134d007f6026e7f9009b83147cb3b600f4b9a100/x/mongo/driver/topology/rtt_monitor.go#L162

Perhaps after disconnect, `ticker` or `r.ctx` are not correctly closed. None of
these connections are ever closed and they keep sending commands to Mongo
servers.
