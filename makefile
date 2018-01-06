test:
	go run simple-wrk.go http://localhost:8080

install: 
	go get github.com/gin-gonic/gin