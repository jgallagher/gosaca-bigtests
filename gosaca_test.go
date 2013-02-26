package gosaca_bigtests

import (
	"fmt"
	"github.com/jgallagher/gosaca"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

var fileCache = map[string][]byte {}

func checkCorrectSuffixArrayBwt(input []byte, SA []int) error {
	n := len(input)

	fmt.Printf("%s: starting sanity check of SA values\n", time.Now())
	// first make sure every element in SA is unique and valid
	indicesSeen := make([]bool, n)
	for i, s := range SA {
		if s < 0 || s >= n {
			return fmt.Errorf("Invalid SA entry: SA[%d] = %d\n", i, s)
		}
		if indicesSeen[s] == true {
			return fmt.Errorf("Duplicate SA entry: SA[%d] = %d was seen before\n", i, s)
		}
		indicesSeen[s] = true
	}

	// Doing a naive check (like the gosaca package does) is way too expensive
	// for these large tests. Instead, we'll compute the Inverse Burrows-Wheel
	// Transform and make sure it matches the original input. The algorithm we
	// follow here is from section 4.2 the original paper (currently available
	// at http://www.hpl.hp.com/techreports/Compaq-DEC/SRC-RR-124.pdf), with
	// the added wrinkle that we need to account for a sentinel character at
	// the end of our string. To deal with this (simply), we make the alphabet
	// size 257 and use -1 as the sentinel for the purposes of the Inv BWT.
	fmt.Printf("%s: starting inverse BWT check\n", time.Now())
	bwtPos := 0
	L := make([]int, n+1)
	//ibwt := make([]int, n+1)
	C := make(map[int]int) // storage for 257 alphabet chars (-1=sentinel, 0-256=data)
	P := make([]int, n+1)
	// construct bwt from SA
	L[0] = int(input[n-1])
	for i := 0; i < n; i++ {
		if SA[i] == 0 {
			bwtPos = i + 1
			L[i+1] = -1
		} else {
			L[i+1] = int(input[SA[i]-1])
		}
	}
	for i := 0; i < n+1; i++ {
		P[i] = C[L[i]]
		C[L[i]]++
	}
	sum := 0
	for i := -1; i < 256; i++ {
		sum += C[i]
		C[i] = sum - C[i]
	}

	// now step through L, comparing each character of the inverse BWT (which
	// we build up from right-to-left) to the original input.
	if L[bwtPos] != -1 {
		return fmt.Errorf("Inverse BWT did not end with sentinel")
	}
	for i := n - 1; i >= 0; i-- {
		bwtPos = P[bwtPos] + C[L[bwtPos]]
		if L[bwtPos] != int(input[i]) {
			return fmt.Errorf("Inverse BWT did not produce original string: position %d: IBWT=%d, input=%d", i, L[bwtPos], input[i])
		}
	}
	return nil
}

func checkSaOfFile(t *testing.T, ws *gosaca.WorkSpace, filename string) {
	fh, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()

	b, err := ioutil.ReadAll(fh)
	if err != nil {
		t.Fatal(err)
	}

	SA := make([]int, len(b))
	fmt.Printf("%s: starting SA on %s\n", time.Now(), filename)
	ws.ComputeSuffixArray(b, SA)
	if err := checkCorrectSuffixArrayBwt(b, SA); err != nil {
		t.Fatalf("bwt check failure on file %s: %s", filename, err)
	}
}

func TestLargeFiles(t *testing.T) {
	ws := &gosaca.WorkSpace{}
	for _, filename := range []string{
		path.Join("large_corpus", "chr22.dna"),
		path.Join("large_corpus", "etext99"),
		path.Join("large_corpus", "gcc-3.0.tar"),
		path.Join("large_corpus", "howto"),
		path.Join("large_corpus", "jdk13c"),
		path.Join("large_corpus", "linux-2.4.5.tar"),
		path.Join("large_corpus", "rctail96"),
		path.Join("large_corpus", "rfc"),
		path.Join("large_corpus", "sprot34.dat"),
		path.Join("large_corpus", "w3c2"),
	} {
		checkSaOfFile(t, ws, filename)
	}
}

func TestGauntletFiles(t *testing.T) {
	ws := &gosaca.WorkSpace{}
	for _, filename := range []string{
		path.Join("gauntlet_corpus", "abac"),
		path.Join("gauntlet_corpus", "abba"),
		path.Join("gauntlet_corpus", "book1x20"),
		path.Join("gauntlet_corpus", "fib_s14930352"),
		path.Join("gauntlet_corpus", "fss10"),
		path.Join("gauntlet_corpus", "fss9"),
		path.Join("gauntlet_corpus", "houston"),
		path.Join("gauntlet_corpus", "paper5x80"),
		path.Join("gauntlet_corpus", "test1"),
		path.Join("gauntlet_corpus", "test2"),
		path.Join("gauntlet_corpus", "test3"),
	} {
		checkSaOfFile(t, ws, filename)
	}
}

func runBenchmark(b *testing.B, filename string) {
	b.StopTimer()
	ws := &gosaca.WorkSpace{}
	input := fileCache[filename]
	if input == nil {
		fh, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		input, err = ioutil.ReadAll(fh)
		fh.Close()
		if err != nil {
			panic(err)
		}
		fileCache[filename] = input
	}
	SA := make([]int, len(input))
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		ws.ComputeSuffixArray(input, SA)
	}
}

func Benchmark_chr22dna(b *testing.B) { runBenchmark(b, path.Join("large_corpus", "chr22.dna")) }
func Benchmark_etext99(b *testing.B)  { runBenchmark(b, path.Join("large_corpus", "etext99")) }
func Benchmark_gcc30tar(b *testing.B) { runBenchmark(b, path.Join("large_corpus", "gcc-3.0.tar")) }
func Benchmark_howto(b *testing.B)    { runBenchmark(b, path.Join("large_corpus", "howto")) }
func Benchmark_jdk13c(b *testing.B)   { runBenchmark(b, path.Join("large_corpus", "jdk13c")) }
func Benchmark_linux245tar(b *testing.B) {
	runBenchmark(b, path.Join("large_corpus", "linux-2.4.5.tar"))
}
func Benchmark_rctail96(b *testing.B)   { runBenchmark(b, path.Join("large_corpus", "rctail96")) }
func Benchmark_rfc(b *testing.B)        { runBenchmark(b, path.Join("large_corpus", "rfc")) }
func Benchmark_sprot34dat(b *testing.B) { runBenchmark(b, path.Join("large_corpus", "sprot34.dat")) }
func Benchmark_w3c2(b *testing.B)       { runBenchmark(b, path.Join("large_corpus", "w3c2")) }
func Benchmark_abac(b *testing.B)       { runBenchmark(b, path.Join("gauntlet_corpus", "abac")) }
func Benchmark_abba(b *testing.B)       { runBenchmark(b, path.Join("gauntlet_corpus", "abba")) }
func Benchmark_book1x20(b *testing.B)   { runBenchmark(b, path.Join("gauntlet_corpus", "book1x20")) }
func Benchmark_fib_s14930352(b *testing.B) {
	runBenchmark(b, path.Join("gauntlet_corpus", "fib_s14930352"))
}
func Benchmark_fss10(b *testing.B)     { runBenchmark(b, path.Join("gauntlet_corpus", "fss10")) }
func Benchmark_fss9(b *testing.B)      { runBenchmark(b, path.Join("gauntlet_corpus", "fss9")) }
func Benchmark_houston(b *testing.B)   { runBenchmark(b, path.Join("gauntlet_corpus", "houston")) }
func Benchmark_paper5x80(b *testing.B) { runBenchmark(b, path.Join("gauntlet_corpus", "paper5x80")) }
func Benchmark_test1(b *testing.B)     { runBenchmark(b, path.Join("gauntlet_corpus", "test1")) }
func Benchmark_test2(b *testing.B)     { runBenchmark(b, path.Join("gauntlet_corpus", "test2")) }
func Benchmark_test3(b *testing.B)     { runBenchmark(b, path.Join("gauntlet_corpus", "test3")) }
