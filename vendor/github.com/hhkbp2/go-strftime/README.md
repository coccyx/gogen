go-strftime
===========

```go-strftime``` is a Golang library that implements the Python-like Time Format Function [```strftime()```][python-stftime].

According to this discussion [Issue 444: Implement strftime][issue-444], in the eyes of the Golang designer, the ```strftime()``` is a bad interface and should be replaced by [```time.Format()```][time-format-golang] because the later one has a simpler format representation. However the formats could be used in ```time.Format()``` is strictly limited to a few pre-defined constants. Sometimes it would become error-prone and tricky to use the format right to produce what format programmers want exactly other than the pre-defineds.

So this library is created to provide a convenient interface for those who like to use ```strftime()``` in Golang.

### Usage

Get the code down using the standard go tool:

```bash
go get github.com/hhkbp2/go-strftime
```

and write your code like

```go
import "github.com/hhkbp2/go-strftime"

formatted := stftime.Format("%Y-%m-%d %H:%M:%S %9n", time.Now())
```

Please refer to the Python [```strftime``` doc][python-stftime] for all the supported format strings. Note that there are two differences: 

1. %c returns RFC1123 which is a bit different from what Python does  
2. %[1-9]n is added for fractional second. e.g. "%3n" means to print millisecond offset with this second. "%6n" means to print microsecond. "%9n" is for nanosecond.

### Contributors

This library is originally developed by [Miki Tebeka][tebeka-page] and forked by [Dylan Wen][dylan-page] to add fractional format support.

The original repository could be found at [https://bitbucket.org/tebeka/strftime][original-repo].


[python-stftime]: http://docs.python.org/2/library/time.html#time.strftime

[issue-444]: https://code.google.com/p/go/issues/detail?id=444

[time-format-golang]: http://golang.org/pkg/time/#Time.Format

[tebeka-page]: https://bitbucket.org/tebeka

[dylan-page]: https://github.com/hhkbp2

[original-repo]: https://bitbucket.org/tebeka/strftime
