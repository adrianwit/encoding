/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package benchtools

import (
	"bytes"
	"compress/gzip"
	"compress/lzw"
	"fmt"
	"github.com/golang/snappy"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/adrianwit/encoding"
	"github.com/adrianwit/encoding/cursor"
)

func TestCodec(codec encoding.Integer, in []int32, sizes []int) {
	for _, k := range sizes {
		if k > len(in) {
			continue
		}

		dur, out, err := Compress(codec, in[:k], k)
		if err != nil {
			log.Fatal(err)
		}

		dur2, out2, err2 := Uncompress(codec, out, k)
		if err2 != nil {
			log.Fatal(err2)
		}

		//log.Printf("benchtools/TestCodec: %f %.2f %.2f\n", float64(len(out)*32)/float64(k), (float64(k)/(float64(dur)/1000000000.0)/1000000.0), (float64(k)/(float64(dur2)/1000000000.0)/1000000.0))
		fmt.Printf("%f %.2f %.2f\n", float64(len(out)*32)/float64(k), (float64(k) / (float64(dur) / 1000000000.0) / 1000000.0), (float64(k) / (float64(dur2) / 1000000000.0) / 1000000.0))

		for i := 0; i < k; i++ {
			if in[i] != out2[i] {
				log.Fatalf("benchtools/TestCodec: Problem recovering. index = %d, in = %d, recovered = %d, original length = %d, recovered length = %d\n", i, in[i], out2[i], k, len(out2))
			}
		}
	}
}

func PprofCodec(codec encoding.Integer, in []int32, sizes []int) {
	for _, k := range sizes {
		if k > len(in) {
			continue
		}

		dur, out, err := PprofCompress(codec, in[:k], k)
		if err != nil {
			log.Fatal(err)
		}

		dur2, out2, err2 := PprofUncompress(codec, out, k)
		if err2 != nil {
			log.Fatal(err2)
		}

		log.Printf("benchtools/PprofCodec: %f %.2f %.2f\n", float64(len(out)*32)/float64(k), (float64(k) / (float64(dur) / 1000000000.0) / 1000000.0), (float64(k) / (float64(dur2) / 1000000000.0) / 1000000.0))

		for i := 0; i < k; i++ {
			if in[i] != out2[i] {
				log.Fatalf("benchtools/PprofCodec: Problem recovering. index = %d, in = %d, recovered = %d, original length = %d, recovered length = %d\n", i, in[i], out2[i], k, len(out2))
			}
		}
	}
}

func Compress(codec encoding.Integer, in []int32, length int) (duration int64, out []int32, err error) {
	return RunCompress(codec, in, length, false)
}

func Uncompress(codec encoding.Integer, in []int32, length int) (duration int64, out []int32, err error) {
	return RunUncompress(codec, in, length, false)
}

func PprofCompress(codec encoding.Integer, in []int32, length int) (duration int64, out []int32, err error) {
	return RunCompress(codec, in, length, true)
}

func PprofUncompress(codec encoding.Integer, in []int32, length int) (duration int64, out []int32, err error) {
	return RunUncompress(codec, in, length, true)
}

func RunCompress(codec encoding.Integer, in []int32, length int, prof bool) (duration int64, out []int32, err error) {
	out = make([]int32, length*2)
	inpos := cursor.New()
	outpos := cursor.New()

	now := time.Now()
	if prof {
		f, e := os.Create("cpu.compress.pprof")
		if e != nil {
			log.Fatal(e)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
	}

	if err = codec.Compress(in, inpos, len(in), out, outpos); err != nil {
		return 0, nil, err
	}
	since := time.Since(now).Nanoseconds()

	if prof {
		pprof.StopCPUProfile()
	}

	return since, out[:outpos.Get()], nil
}

func RunUncompress(codec encoding.Integer, in []int32, length int, prof bool) (duration int64, out []int32, err error) {
	out = make([]int32, length)
	inpos := cursor.New()
	outpos := cursor.New()

	if prof {
		f, e := os.Create("cpu.uncompress.pprof")
		if e != nil {
			log.Fatal(e)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
	}

	now := time.Now()
	if err = codec.Uncompress(in, inpos, len(in), out, outpos); err != nil {
		return 0, nil, err
	}
	since := time.Since(now).Nanoseconds()

	if prof {
		pprof.StopCPUProfile()
	}

	return since, out[:outpos.Get()], nil
}

func RunTestGzip(data []byte) {
	log.Printf("encoding/RunTestGzip: Testing comprssion Gzip\n")

	var compressed bytes.Buffer
	w := gzip.NewWriter(&compressed)
	defer w.Close()
	now := time.Now()
	w.Write(data)

	cl := compressed.Len()
	log.Printf("encoding/RunTestGzip: Compressed from %d bytes to %d bytes in %d ns\n", len(data), cl, time.Since(now).Nanoseconds())

	recovered := make([]byte, len(data))
	r, _ := gzip.NewReader(&compressed)
	defer r.Close()

	total := 0
	n := 100
	var err error = nil
	for err != io.EOF && n != 0 {
		n, err = r.Read(recovered[total:])
		total += n
	}
	log.Printf("encoding/RunTestGzip: Uncompressed from %d bytes to %d bytes in %d ns\n", cl, len(recovered), time.Since(now).Nanoseconds())
}

func RunTestLZW(data []byte) {
	log.Printf("encoding/RunTestLZW: Testing comprssion LZW\n")

	var compressed bytes.Buffer
	w := lzw.NewWriter(&compressed, lzw.MSB, 8)
	defer w.Close()
	now := time.Now()
	w.Write(data)

	cl := compressed.Len()
	log.Printf("encoding/RunTestLZW: Compressed from %d bytes to %d bytes in %d ns\n", len(data), cl, time.Since(now).Nanoseconds())

	recovered := make([]byte, len(data))
	r := lzw.NewReader(&compressed, lzw.MSB, 8)
	defer r.Close()

	total := 0
	n := 100
	var err error = nil
	for err != io.EOF && n != 0 {
		n, err = r.Read(recovered[total:])
		total += n
	}
	log.Printf("encoding/RunTestLZW: Uncompressed from %d bytes to %d bytes in %d ns\n", cl, len(recovered), time.Since(now).Nanoseconds())
}

func RunTestSnappy(data []byte) {
	log.Printf("encoding/RunTestSnappy: Testing comprssion Snappy\n")

	now := time.Now()
	e := snappy.Encode(nil, data)

	log.Printf("encoding/RunTestSnappy: Compressed from %d bytes to %d bytes in %d ns\n", len(data), len(e), time.Since(now).Nanoseconds())

	d, err := snappy.Decode(nil, e)
	if err != nil {
		log.Fatalf("encoding/RunTestSnappy: decoding error: %v\n", err)
	}
	log.Printf("encoding/RunTestSnappy: Uncompressed from %d bytes to %d bytes in %d ns\n", len(e), len(d), time.Since(now).Nanoseconds())

	if !bytes.Equal(data, d) {
		log.Fatalf("encoding/RunTestSnappy: roundtrip mismatch\n")
	}
}
