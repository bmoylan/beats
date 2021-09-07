package readfile

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"

	"golang.org/x/text/encoding"
)

func BenchmarkEncoderReader(b *testing.B) {
	const (
		bufferSize   = 1024
		lineMaxLimit = 1000000 // never hit by the input data
	)

	runBench := func(name string, lineMaxLimit int, lines []byte) {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for bN := 0; bN < b.N; bN++ {
				reader, err := NewEncodeReader(ioutil.NopCloser(bytes.NewReader(lines)), Config{encoding.Nop, bufferSize, LineFeed, lineMaxLimit})
				if err != nil {
					b.Fatal("failed to initialize reader:", err)
				}
				// Read decodec lines and test
				size := 0
				for i := 0; ; i++ {
					msg, err := reader.Next()
					if err != nil {
						if err == io.EOF {
							b.ReportMetric(float64(i), "processed_lines")
							break
						} else {
							b.Fatal("unexpected error:", err)
						}
					}
					size += msg.Bytes
				}
				b.ReportMetric(float64(size), "processed_bytes")
			}
		})
	}

	runBench("buffer-sized lines", lineMaxLimit, createBenchmarkLines(100, 1020))
	runBench("short lines", lineMaxLimit, createBenchmarkLines(100, 10))
	runBench("long lines", lineMaxLimit, createBenchmarkLines(100, 10_000))
	// short lineMaxLimit to exercise skipUntilNewLine
	runBench("skip lines", 1024, createBenchmarkLines(100, 10_000))
}

func createBenchmarkLines(numLines int, lineLength int) []byte {
	buf := bytes.NewBuffer(nil)
	for i := 0; i < numLines; i++ {
		line := make([]byte, hex.DecodedLen(lineLength))
		if _, err := rand.Read(line); err != nil {
			panic(fmt.Sprintf("failed to generate random input: %v", err))
		}
		buf.WriteString(hex.EncodeToString(line))
		buf.WriteRune('\n')
	}
	return buf.Bytes()
}
