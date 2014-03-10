# serial port test

group = 
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

console.log JSON.stringify group, null, 4
