# simple-wrk

Another http test framework with golang, simple is more.

## Installation

Use `go get github.com/binatify/simple-wrk`


## Usage

```
$ simple-wrk -c 200 -d 5 http://localhost:8080

Running 5s test @ http://localhost:8080
56539 requests in 4.964309804s, 6.58MB read
Requests/sec:		11389.10
Transfer/sec:		1.33MB
Avg Req Time:		878.032µs
Fastest Request:	139.482µs
Slowest Request:	985.130606ms
Number of Errors:	0
```

## Architecture

```
                                    --------
                                   | Loader |
                                    --------
                                       | go
                                       V
                         --------             --------
                        | Client |    ....   | Client |  
                         --------             --------
                                       | channel                         
                                       V
                                    --------    
                                   | Summary | 
                                    --------  
                                       
```
