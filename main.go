package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type task struct {
	srcDB   int
	srcSec  int
	destDB  int
	destSec int
}

var (
	h bool
	v bool
	i string
	t string
)

func init() {
	flag.BoolVar(&h, "h", false, "help")
	flag.BoolVar(&v, "v", false, "version")
	flag.StringVar(&i, "i", "./eeprom.log", "path to log file")
	flag.StringVar(&t, "t", "./task.yml", "path to task file")
}

func getTask() {

}
func exchange(srcDB, srcSec, destDB, destSec int, pathToLog string) {
	if srcDB < 0 || srcDB > 14 {
		return
	} else if srcSec < 0 || srcSec > 29 {
		return
	}
	eeprom, _ := os.Open(pathToLog)
	var djm []string
	br := bufio.NewReader(eeprom)
	for {
		line, err := br.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			break
		}
		djm = append(djm, string(line))
	}
	srcLine := 544*srcDB + srcSec + 3
	if srcSec > 13 {
		srcLine++
	}
	destLine := 544*destDB + destSec + 3
	if destLine > 13 {
		destLine++
	}
	djm[srcLine], djm[destLine] = djm[destLine], djm[srcLine]
	for i := 0; i < 16; i++ {
		srcLine = 544*srcDB + srcSec*17 + 34 + i
		destLine = 544*destDB + destSec*17 + 34 + i
		djm[srcLine], djm[destLine] = djm[destLine], djm[srcLine]
	}
	//ioutil.WriteFile("output.txt", []byte(strings.Join(djm, "")), os.ModeAppend)
	ioutil.WriteFile("output.txt", []byte("dddd"), os.ModeAppend)
	//fmt.Print(len(djm))
}

func main() {
	flag.Parse()
	if v {
		fmt.Println("eeprom sector exchange tool(eset) V0.1.0")
	}
	if h {
		fmt.Fprintf(os.Stdout, `eeprom sector exchange tool(eset) version: eset/V0.1.0
Usage: eset [-hv] [-i filename] [-t filename]
		
Options:
`)
		flag.PrintDefaults()
	}
	exchange(-1, 14, 1, 14, "test.txt")
}
