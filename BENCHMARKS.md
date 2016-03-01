## Performance Benchmarks

There are three benchmarks in `benchmark_test.go` that compare the performance of `go-restructure` to that of the standard library `regexp` package. `go-restructure` uses a very slightly modified version of the `regexp` package so the performance of the core regular expression evaluator is very similar; most of the difference is therefore associated with the overhead of reflection.

These benchmarks were computed using `go test -bench=.` on an 2.8 GHz Intel Core i7 processor running OSX 10.10.5.

The first benchmark involves finding the first floating point number in a string of a few thousand characters. `go-restructure` takes around 8% longer than the standard library:

```
go-restructure		32428 ns/op
stdlib/regexp		30060 ns/op
```

The second benchmark involves parsing a short email address. `go-restructure` takes around 
40% longer than the standard library:

```
go-restructure		1188 ns/op
stdlib/regexp		844 ns/op
```

The third benchmark involves finding all python import statements in a file of around one hundred lines of python source. `go-restructure` takes around 2x longer than the standard library:

```
go-restructure		695 ns/op
stdlib/regexp		337 ns/op
```

The high overhead for `go-restructure` on the last benchmark is probably due to `go-restructure` allocating a struct to hold the results of each match found by `FindAll`. In most cases this performance overhead will be a small price to pay for composable, inspectable regular expressions, particularly when it amonuts to the difference between one third of a microsecond and two thirds of a microsecond. However, applications that execute a very large number of regular expressions for which performance is critical may be well advised to use the standard library `regexp` package directly.
