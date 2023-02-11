# horizon node

currently prototype

## how to run 

simple run example and connect two nodes

```go run . --config config.json```

and in a second terminal
```go run . --config config2.json```

currently these nodes will exchange messages about the state. dont produce new state. this requires a simulator passing transactions to them.