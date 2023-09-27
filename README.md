[![Go Report](https://goreportcard.com/badge/github.com/diegostamigni/logzzap)](https://goreportcard.com/report/github.com/diegostamigni/logzzap)

# Logz Zap

A simple zapcore.Core implementation to integrate with Logz.

To use, initialize logz like normal, create a new LogzCore, then wrap with a NewTee. [See the example code](example/main.go) for a detailed example.

## Testing 

To test this code use `LZ_TOKEN=MY_LOGZ_TOKEN go test`
