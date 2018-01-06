package util

import (
	"fmt"
	"net/http"
	"time"
)

const (
	Byte  = 1
	KByte = Byte * 1024
	MByte = KByte * 1024
	GByte = MByte * 1024
)

type ByteSize float64

func (this ByteSize) String() string {
	var (
		rt     ByteSize
		suffix string
	)

	if this > GByte {
		rt = this / GByte
		suffix = "GB"
	} else if this > MByte {
		rt = this / MByte
		suffix = "MB"
	} else if this > KByte {
		rt = this / KByte
		suffix = "KB"
	} else {
		rt = this
		suffix = "bytes"
	}

	return fmt.Sprintf("%.2f%v", rt, suffix)
}

func MaxDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 > d2 {
		return d1
	}
	return d2
}

func MinDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}

func EstimateHttpHeadersSize(headers http.Header) (result int64) {
	result = 0

	for k, v := range headers {
		result += int64(len(k) + len(": \r\n"))
		for _, s := range v {
			result += int64(len(s))
		}
	}

	result += int64(len("\r\n"))

	return result
}
