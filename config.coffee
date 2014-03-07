group = 
  workers: [
    { name: "lr", type: "LogReader" }
    { name: "rf", type: "Pipe" } # used to inject an "[RF12demo...]" line
    { name: "w1", type: "LogReplayer" }
    # { name: "w1", type: "SerialIn" }
    { name: "ts", type: "TimeStamp" }
    { name: "st", type: "SketchType" }
    { name: "nm", type: "NodeMap" }
    { name: "hp", type: "Node-homePower" }
    { name: "or", type: "Node-ookRelay" }
    { name: "dp", type: "Dispatcher" }
    { name: "rb", type: "Node-radioBlip" }
    { name: "rn", type: "Node-roomNode" }
    { name: "p", type: "Printer" }
  ]
  connections: [
    { from: "lr.Out", to: "w1.In" }
    { from: "rf.Out", to: "ts.In" }
    { from: "w1.Out", to: "ts.In" }
    { from: "ts.Out", to: "st.In" }
    { from: "st.Out", to: "nm.In" }
    { from: "nm.Out", to: "hp.In" }
    { from: "hp.Out", to: "or.In" }
    { from: "or.Out", to: "dp.In" }
    { from: "or.Type", to: "dp.Use" }
    { from: "dp.Out", to: "rb.In" }
    { from: "rb.Out", to: "rn.In" }
    { from: "rn.Out", to: "p.In" }
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
    # { data: "/dev/tty.usbserial-A901ROSM", to: "w1.Port" }
  ]

console.log JSON.stringify group, null, 4
