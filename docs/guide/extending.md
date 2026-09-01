---
title: Extending — add a receiver
type: guide
updated: 2026-09-01
---

# Adding a receiver (R10)

A new protocol is added by implementing the documented receiver contract and
registering it — with no changes to Servienta core. The `reference` receiver is
the worked example (`internal/receiver/reference`).

## The contract

```go
type Receiver interface {
    Name() string          // service name, e.g. "myproto"
    Endpoints() []string   // labels this receiver binds; single-surface returns {Name()}
    Start(ctx context.Context, addrs map[string]string, rec Recorder) (map[string]net.Addr, error)
}
```

`Recorder` is your only write path into Servienta:
```go
type Recorder interface {
    Record(service, sourceIP string, content map[string]any) error
    Mode(service string) (mode string, delayMs int)      // honor R9 failure modes
    Response(service string) map[string]any               // read R8 control (request-response only)
}
```

## Steps

1. Create `internal/receiver/myproto/myproto.go` implementing `Receiver`.
2. In `Start`, bind your listener(s) to the given address(es), serve until
   `ctx` is done, and for each message call `rec.Record(Name(), sourceIP, ...)`.
3. Honor `rec.Mode(Name())` for the R9 modes you support (at least `drop`).
4. Register it in `internal/app/app.go` (`receivers(cfg)`), map its endpoint
   label(s) to config addresses, and give it a stand id if it is licensable.
5. Add an acceptance test that sends to it and reads it back via R4.

That is all — reset, run attribution, and fault modes come from the core. The
receiver appears in `/received`, obeys `POST /reset`, and needs no core edits.
