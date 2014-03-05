nodeMap =
  "RF868-5-4": "roomNode"
  
group = 
  workers: [
    { name: "si", type: "SerialIn" }
    { name: "ts", type: "TimeStamp" }
    { name: "st", type: "SketchType" }
    # { name: "nm", type: "Decoder-NodeMap" }
    { name: "nm", type: "Pipe" }
    { name: "p", type: "Printer" }
  ]
  connections: [
    { from: "si.Out", to: "ts.In" }
    { from: "ts.Out", to: "st.In" }
    { from: "st.Out", to: "nm.In" }
    { from: "nm.Out", to: "p.In" }
  ]
  requests: [
    { data: "/dev/tty.usbserial-A901ROSM", to: "si.Port" }
    # { data: nodeMap, to: "nm.Info" }
  ]

console.log JSON.stringify group, null, 4






# group = 
#   workers: [
#     { type: "SerialIn", name: "s" }
#     { type: "TimeStamp", name: "t" }
#     { type: "SketchType", name: "u" }
#     { type: "Printer", name: "p" }
#   ]
#   connections: [
#     { from: "s.Out", to: "t.In" }
#     { from: "t.Out", to: "u.In" }
#     { from: "u.Out", to: "p.In" }
#   ]
#   requests: [
#     data: "/dev/tty.usbserial-A901ROSM", to: "s.Port"
#   ]
# 
# console.log JSON.stringify group, null, 4
