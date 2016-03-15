package main
import (
    "os/exec"
	"fmt"
	"log"
	"os"
	"time"
	"strconv"
)

var usage = `Usage: generator [options...] <url>

Options:
  -t  Type of Workloads, 1: Circule 2: Burst 
  -h  Custom Web Server Address, name1:value1
`

var clis int64
var ratefile *os.File 

func UsageAndExit() {
	fmt.Fprint(os.Stderr, usage)
	os.Exit(1)
}

func reqRate() {
	//rate := int64(1 / 0.03) * clis 	
	rate := 33 * clis
	ticker1 := time.NewTicker(time.Second * 1)
	for _ = range ticker1.C {
		//log.Print("Request rate: ", rate)	
		//_ = exec.Command("echo", strconv.FormatInt(int64(rate), 10),  ">", "/Users/maxwell/mworks/4paper/reqrate")
		ratefile.WriteString(strconv.FormatInt(rate, 10) + "\n")
	}
}

func cirGen() {
	f, err := os.Open("cfile")
	if err != nil {
		log.Println(err)	
	}
	var i int64    
    for i = 0; i < 5; i++ {
		num := make([]byte, 5)
		_, err = f.Seek(i * 10, 0)
		if err != nil {
			log.Println(err)	
		}
		_, err = f.Read(num)
		if err != nil {
			log.Println(err)	
		}

		cli := make([]byte, 3)
		_, err = f.Seek(i * 10 + 6, 0)
		if err != nil {
			log.Println(err)	
		}
		_, err = f.Read(cli)
		if err != nil {
			log.Println(err)		
		}

		ent := make([]byte, 1)
		_, err = f.Seek(i * 10 + 9, 0)
		_, err = f.Read(ent)
		
		cmd := exec.Command("./myboom/boom/myboomfile", "-n", string(num), "-c", string(cli), os.Args[4])
		clis, _ = strconv.ParseInt(string(cli), 10, 64)
		bytes, err := cmd.Output()
		if err != nil {
       		log.Println("error: ")	
		}
		fmt.Println(string(bytes)) //boom report
	} 
}

func burGen() {
	cmd := exec.Command("./myboom/boom/myboomfile", "-n", "2000", "-c", "200", os.Args[4])
	bytes, err := cmd.Output()
	if err != nil {
       	fmt.Println("error: ")	
	}
	fmt.Println(string(bytes))
}

func main() {
    if len(os.Args) != 5 {
		UsageAndExit()
	}else {
		clis = 0
		ratefile, _ = os.Create("rate")
		go reqRate() 
		if(os.Args[2] == "1") {
			cirGen()	
		} else if(os.Args[2] == "2") {
			burGen()	
		} else {
			UsageAndExit()	
		}
		defer ratefile.Close()
	}
}

