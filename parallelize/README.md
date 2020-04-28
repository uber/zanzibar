# parallelize-go
Parallelize functions in golang

This library works in few modes: a) for doing short io bound work default, it spawns 2x goroutines of CPU. b) If its unsuitable parallel count can be specified. c) for running long-running io bound work provides a function to spawns unbounded goroutine.

It returns early with the first error it sees. note, first error it sees from parallel goroutine and not first error as in parallel many goroutines may simultaneously return an error.
