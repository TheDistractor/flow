circuits = {}

circuits.main =
  gadgets: [
    { name: "p", type: "Printer" }
  ]
  wires: [
    # nothing to connect...
  ]
  feeds: [
    { data: "blah", to: "p.In" }
  ]

console.log JSON.stringify circuits, null, 4
