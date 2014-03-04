group = 
  workers: [
    { type: "SerialIn", name: "s" }
    { type: "SketchType", name: "t" }
    { type: "Printer", name: "p" }
  ]
  connections: [
    { from: "s.Out", to: "t.In" }
    { from: "t.Out", to: "p.In" }
  ]
  requests: [
    data: "/dev/tty.usbserial-A900ad5m", to: "s.Port"
  ]

console.log JSON.stringify group, null, 4
