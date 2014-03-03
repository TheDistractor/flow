group = 
  workers: [
    { type: "SerialIn", name: "jeelink1" }
    { type: "SketchType", name: "jeelink2" }
    { type: "Printer", name: "printer" }
  ]
  connections: [
    { from: "jeelink1.Out", to: "jeelink2.In" }
    { from: "jeelink2.Out", to: "printer.In" }
  ]
  requests: [
    data: "/dev/tty.usbserial-A900ad5m", to: "jeelink1.Port"
  ]

console.log JSON.stringify group, null, 4
