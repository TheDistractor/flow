groups = {}

# default app runs a replay simulation with dynamic decoders
groups.main =
  workers: [
    { name: "lr", type: "LogReader" }
    { name: "rf", type: "Pipe" } # used to inject an "[RF12demo...]" line
    { name: "w1", type: "LogReplayer" }
    { name: "ts", type: "TimeStamp" }
    { name: "fo", type: "FanOut" }
    { name: "lg", type: "Logger" }
    { name: "st", type: "SketchType" }
    { name: "d1", type: "Dispatcher" }
    { name: "nm", type: "NodeMap" }
    { name: "d2", type: "Dispatcher" }
    { name: "p", type: "Printer" }
  ]
  connections: [
    { from: "lr.Out", to: "w1.In" }
    { from: "rf.Out", to: "ts.In" }
    { from: "w1.Out", to: "ts.In" }
    { from: "ts.Out", to: "fo.In" }
    { from: "fo.Out:lg", to: "lg.In" }
    { from: "fo.Out:st", to: "st.In" }
    { from: "st.Out", to: "d1.In" }
    { from: "d1.Out", to: "nm.In" }
    { from: "nm.Out", to: "d2.In" }
    { from: "d2.Out", to: "p.In" }
  ]
  requests: [
    { data: "RFg5i2 roomNode",   to: "nm.Info" }
    { data: "RFg5i3 radioBlip",  to: "nm.Info" }
    { data: "RFg5i4 roomNode",   to: "nm.Info" }
    { data: "RFg5i5 roomNode",   to: "nm.Info" }
    { data: "RFg5i6 roomNode",   to: "nm.Info" }
    { data: "RFg5i9 homePower",  to: "nm.Info" }
    { data: "RFg5i10 roomNode",  to: "nm.Info" }
    { data: "RFg5i11 roomNode",  to: "nm.Info" }
    { data: "RFg5i12 roomNode",  to: "nm.Info" }
    { data: "RFg5i13 roomNode",  to: "nm.Info" }
    { data: "RFg5i14 otRelay",   to: "nm.Info" }
    { data: "RFg5i15 smaRelay",  to: "nm.Info" }
    { data: "RFg5i18 p1scanner", to: "nm.Info" }
    { data: "RFg5i19 ookRelay",  to: "nm.Info" }
    { data: "RFg5i23 roomNode",  to: "nm.Info" }
    { data: "RFg5i24 roomNode",  to: "nm.Info" }
    
    { data: "[RF12demo.10] _ i31* g5 @ 868 MHz", to: "rf.In" }
    { data: "./rfdata/20121130.txt.gz", to: "lr.Name" }
    { data: "./logger", to: "lg.Dir" }
  ]

# serial port test
groups.serial =
  workers: [
    { name: "sp", type: "SerialPort" }
    { name: "st", type: "SketchType" }
    { name: "d1", type: "Dispatcher" }
    { name: "nm", type: "NodeMap" }
    { name: "d2", type: "Dispatcher" }
  ]
  connections: [
    { from: "sp.From", to: "st.In" }
    { from: "st.Out", to: "d1.In" }
    { from: "d1.Out", to: "nm.In" }
    { from: "nm.Out", to: "d2.In" }
  ]
  requests: [
    { data: "RFg5i3 radioBlip",  to: "nm.Info" }
    { data: "RFg5i9 homePower",  to: "nm.Info" }
    { data: "RFg5i13 roomNode",  to: "nm.Info" }
    { data: "RFg5i14 otRelay",   to: "nm.Info" }
    { data: "RFg5i15 smaRelay",  to: "nm.Info" }
    { data: "RFg5i18 p1scanner", to: "nm.Info" }
    { data: "RFg5i19 ookRelay",  to: "nm.Info" }
    
    { data: "/dev/tty.usbserial-A901ROSM", to: "sp.Port" }
  ]

# simple jeebus setup, with dummy websocket support
groups.jeebus =
  workers: [
    { name: "http", type: "HttpServer" }
  ]
  requests: [
    { tag: "/", data: "../jeebus/app",  to: "http.Handlers" }
    { tag: "/base/", data: "../jeebus/base",  to: "http.Handlers" }
    { tag: "/common/", data: "../jeebus/common",  to: "http.Handlers" }
    { tag: "/ws", data: "<websocket>",  to: "http.Handlers" }
    { data: ":3000",  to: "http.Start" }
  ]

# define the websocket handler as just a pipe back to the browser for now
groups["WebSocket-jeebus"] =
  workers: [
    { name: "p", type: "Pipe" }
  ]
  mappings: [
    { external: "In", internal: "p.In" }
    { external: "Out", internal: "p.Out" }
  ]

console.log JSON.stringify groups, null, 4
