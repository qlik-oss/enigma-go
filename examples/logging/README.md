# Logging

This example shows how to implement the TrafficLogger interface and log the JSON traffic to `stdout`.

```
Sent: {"jsonrpc":"2.0","delta":false,"method":"CreateSessionApp","handle":-1,"id":1,"params":[]}
Received: {"jsonrpc":"2.0","method":"OnConnected","params":{"qSessionState":"SESSION_CREATED"}}
Received: {"jsonrpc":"2.0","id":1,"result":{"qReturn":{"qType":"Doc","qHandle":1},"qSessionAppId":"SessionApp_9ff821c1-28ca-4b23-9ad5-1f2e6606f475"},"change":[1]}
Sent: {"jsonrpc":"2.0","delta":false,"method":"SetScript","handle":1,"id":2,"params":["\nTempTable:\nLoad\nRecNo() as Field1,\nRand() as Field2,\nRand() as Field3\nAutoGenerate 100\n"]}
Received: {"jsonrpc":"2.0","id":2,"result":{}}
Sent: {"jsonrpc":"2.0","delta":false,"method":"DoReload","handle":1,"id":3,"params":[0,false,false]}
```

## Runnable code

* [Traffic Log](./traffic-log.go)
