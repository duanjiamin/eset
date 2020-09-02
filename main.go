package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/snksoft/crc"
	yaml "gopkg.in/yaml.v2"
)

type task struct {
	SrcDB   int `yaml:"srcdb"`
	SrcSec  int `yaml:"srcsec"`
	DestDB  int `yaml:"destdb"`
	DestSec int `yaml:"destsec"`
}
type settask struct {
	DestDB  int `yaml:"destdb"`
	DestSec int `yaml:"destsec"`
	Count   int `yaml:"count"`
}
type yml struct {
	Exchange []task    `yaml:"exchange"`
	Copy     []task    `yaml:"copy"`
	Set      []settask `yaml:"set"`
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

var conf yml

func getTask() {
	//conf = yml{}
	yamlFile, err := ioutil.ReadFile(t)
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &conf)

	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}

func exchange(srcDB, srcSec, destDB, destSec int, djm []string) {
	if srcDB < 0 || srcDB > 14 {
		return
	} else if srcSec < 0 || srcSec > 29 {
		return
	}
	if destDB < 0 || destDB > 14 {
		return
	} else if destSec < 0 || destSec > 29 {
		return
	}

	srcLine := 544*srcDB + srcSec + 3
	if srcSec > 13 {
		srcLine++
	}
	destLine := 544*destDB + destSec + 3
	if destSec > 13 {
		destLine++
	}
	djm[srcLine], djm[destLine] = djm[destLine], djm[srcLine]

	for i := 0; i < 17; i++ {
		srcLine = 544*srcDB + srcSec*17 + 34 + i
		destLine = 544*destDB + destSec*17 + 34 + i
		djm[srcLine], djm[destLine] = djm[destLine], djm[srcLine]
	}
	//ioutil.WriteFile("output.txt", []byte(strings.Join(djm, "")), os.ModeAppend)
	//ioutil.WriteFile("output.txt", []byte("dddd"), os.ModeAppend)
	//fmt.Print(len(djm))
}

func copySec(srcDB, srcSec, destDB, destSec int, djm []string) {
	if srcDB < 0 || srcDB > 14 {
		return
	} else if srcSec < 0 || srcSec > 29 {
		return
	}
	if destDB < 0 || destDB > 14 {
		return
	} else if destSec < 0 || destSec > 29 {
		return
	}

	srcLine := 544*srcDB + srcSec + 3
	if srcSec > 13 {
		srcLine++
	}
	destLine := 544*destDB + destSec + 3
	if destSec > 13 {
		destLine++
	}
	djm[destLine] = djm[srcLine]
	for i := 0; i < 16; i++ {
		srcLine = 544*srcDB + srcSec*17 + 34 + i
		destLine = 544*destDB + destSec*17 + 34 + i
		djm[destLine] = djm[srcLine]
	}
}

func setSec(destDB, destSec, count int, djm []string) {
	if destDB < 0 || destDB > 14 {
		return
	} else if destSec < 0 || destSec > 29 {
		return
	} else if count > 0xFFFFFFFF {
		return
	}
	secLine := 544*destDB + destSec + 3
	if destSec > 13 {
		secLine++
	}
	countStr := fmt.Sprint(strconv.FormatInt(int64(count), 16))
	tempStr := "00000000"
	if len(countStr) < 8 {
		countStr = tempStr[:8-len(countStr)] + countStr
	}

	//fmt.Println(djm[secLine])

	tempStr = countStr[:2] + " " + countStr[2:4] + " " + countStr[4:6] + " " + countStr[6:8]
	djm[secLine] = djm[secLine][:21] + tempStr + djm[secLine][32:]
	dataForCRC := strings.ReplaceAll(djm[secLine][9:32], " ", "")

	var byteForCRC []byte
	for i := 0; i < 16; i += 2 {
		if v, err := strconv.ParseUint(dataForCRC[i:i+2], 16, 16); err == nil {
			byteForCRC = append(byteForCRC, byte(v))
		}
	}
	//fmt.Println(byteForCRC)
	newCRC := strings.ToUpper(fmt.Sprint(strconv.FormatInt(int64(crc.CalculateCRC(crc.CCITT, byteForCRC)), 16)))
	djm[secLine] = djm[secLine][:33] + newCRC[:2] + " " + newCRC[2:] + djm[secLine][38:]
	//fmt.Println(djm[secLine])
}

func doTask() {
	eeprom, err := os.Open(i)
	if err != nil {
		log.Fatalf("OpenFile: %v", err)
	}
	defer eeprom.Close()
	var djm []string
	br := bufio.NewReader(eeprom)
	for {
		line, err := br.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			break
		}
		djm = append(djm, string(line))
	}
	for _, v := range conf.Exchange {
		exchange(v.SrcDB, v.SrcSec, v.DestDB, v.DestSec, djm)
	}
	for _, v := range conf.Copy {
		copySec(v.SrcDB, v.SrcSec, v.DestDB, v.DestSec, djm)
	}
	for _, v := range conf.Set {
		setSec(v.DestDB, v.DestSec, v.Count, djm)
	}
	if err := ioutil.WriteFile("output_"+fmt.Sprint(time.Now().Unix())+".log", []byte(strings.Join(djm, "")), 0777); err != nil {
		fmt.Println(err)
	}
}

func main() {
	flag.Parse()
	if v {
		fmt.Println("eeprom sector exchange tool(eset) V0.1.0")
	} else if h {
		fmt.Fprintf(os.Stdout, `eeprom sector exchange tool(eset) version: eset/V0.1.0
Usage: eset [-hv] [-i filename] [-t filename]
  
Options:
`)
		flag.PrintDefaults()
	} else {
		getTask()
		doTask()
	}

}
