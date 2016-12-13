# Benchmarks

To run the benchmarks, compile & run the `runner`

```
make bins
./benchmarks/runner/runner
```

This will run the benchmark runner..

Next you can use `wrk` to generate load,
There are multiple lua scripts in `./benchmarks/*.lua`
each with their own instructions.

For example:

```
wrk -t12 -c400 -d30s -s ./benchmarks/contacts_1KB.lua http://localhost:8093/contacts/foo/contacts
```

This will generate load against the gateway.