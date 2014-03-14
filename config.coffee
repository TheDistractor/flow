groups = {}

groups.main =
  workers: [
    { name: "p", type: "Printer" }
  ]
  connections: [
    # nothing to connect...
  ]
  requests: [
    { data: "blah", to: "p.In" }
  ]

console.log JSON.stringify groups, null, 4
