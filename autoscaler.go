package main

import (
	"fmt"
	"log"
	//"sync"
	"os"
	"time"
	"os/exec"
	"strconv"
	"bufio"
)

var usage = ` Usage: autoscaler [options] <url>

Options:

-h Custom Web Cluster Address, name1:value1
-r request rate (sent per second) threshold
-c cpu usage threshold
-t response time (average in one second) threshold
`
var totalServer int
var webAddress string
var kupper float32
var rateupper float32
var ttlupper float32
var ttllower float32

//var modelType chan int
var curTTL chan float32
var curRate chan float32
var proRate chan float32
var scaleType chan int
var monRate chan float32
var numServer chan int
//var kRate chan float32

var hWin []float32
var pWin []float32

func calave(win []float32) (ave float32) {
	var sum float32
	for i := 0; i < len(win) - 1; i++ {
		sum += win[i]
	}
	return sum / float32(len(win) - 1)
}

func monitorttl() {
	go func() {
		cmd := exec.Command("/bin/bash", "-C", "./watchttlave.sh")
		if _, err := cmd.Output(); err != nil {
			log.Println(err)
		}
	} ()
	time.Sleep(time.Duration(1) * time.Second)
    fTTL, err := os.Open("./ttlave")
	if err != nil {
		log.Println(err)
	}
	reader := bufio.NewReader(fTTL)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
/*  ttlWin := make([]float32, 0, 2700)
	var count int64
*/
	var ave float32
	for scanner.Scan() {
		str := scanner.Text()
		n, _ := strconv.ParseFloat(str, 64)
		//ttlWin = append(ttlWin, float32(n))
/*		ttlWin[count] = float32(n)
		count++
		if count == 2700 {
			ave = calave(ttlWin)
			count = 0
			curTTL<-ave
			time.Sleep(time.Duration(freq) * time.Second)
		}
*/
		ave = float32(n)
		curTTL<- ave
	}
}

func monitorrate() {
	time.Sleep(time.Duration(1) * time.Second)
	fRate, err := os.Open("./rate")
	if err != nil {
		log.Println(err)
	}
	reader := bufio.NewReader(fRate)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	rateWin := make([]float32, 1)
	for scanner.Scan() {
		str := scanner.Text()
		n, _ := strconv.ParseFloat(str, 64)
		rateWin[0] = float32(n)
		curRate<- rateWin[0]
//		time.Sleep(time.Duration(freq) * time.Second)
	}
}


func monitor() {
	log.Println("Monitor is Running")
	//var cr float64
	//var ct float64
	go func() {
		monitorttl()
	/*	cmd := exec.Command("./Monitorttl")
		re, err := cmd.Output()
		if err != nil {
			log.Println(err)
		}

		ct, _ = strconv.ParseFloat(re, 32)
		curTTL<- float32(ct)
	*/
	}()
	go func() {
	/*	cmd := exec.Command("./Monitorate")
		re, err := cmd.Output()
		if err != nil {
			log.Println(err)
		}
		cr, _ = strconv.ParseFloat(string(re), 32)
		curRate<- float32(cr)
	*/
		monitorrate()

	}()

}

func Calk() (k float32) {
	//var k float32
/*	var sum float32
	for i := len(pWin) - 1; i > len(pWin) - 4; i-- {
		sum += (pWin[i] - pWin[i - 1]) / pWin[i - 1]
	}
	k = sum / float32(len(pWin) - 1)
*/
	k = (pWin[len(pWin) - 1] - pWin[len(pWin) - 2]) / pWin[len(pWin) - 2]
	return k
	//kRate <- k
}

func CalNum(cr float32) {
	var n int
	rateupper = rateupper * float32(totalServer)
	n = int(cr / rateupper)
	numServer<-n
}

func addToSlice(win []float32, size int, n float32) {
	win = append(win ,n)
	win = win[1: size]
}

func alerter(freq int64) {
	log.Println("Alerter is Running")
	var k float32
	var pr float32
	var cr float32
	var ct float32
	var tpr float32
	var tcr float32
	var tct float32
    var prcnt int64
	var crcnt int64
	var ctcnt int64


	for {
		select {
		//case: k = <-kRate
			case pr = <-proRate: {
				prcnt++
				if prcnt == freq {
					tpr = pr
					prcnt = 0
				}
			}
			case cr = <-curRate: {
				crcnt++
				if crcnt == freq {
					addToSlice(hWin, 10, cr)
					addToSlice(pWin, 4, cr)
					tcr = cr
					crcnt = 0
				}
			}
			case ct = <-curTTL: {
				ctcnt++
				if ctcnt == freq {
					tct = ct
					ctcnt = 0
				}
			}
		}

		//hWin = append(hWin, cr)
		//pWin = append(pWin, cr)

		if tcr > rateupper || tpr > rateupper  || tct > ttlupper {
			CalNum(cr)
			k = Calk()
			if k > kupper {
				scaleType <- 2
			}else {
				scaleType <- 1
			}
		}

		if tct < ttllower {
			scaleType <- 3
		}

	}
}

func modeler(t int) {
	log.Println("Moderler is Running")
	var pr float32
	//t = 1                              // 1: MA  2: AR
	if t == 1 {
		i := len(hWin) - 1
		pr = (hWin[i] + hWin[i - 1]) / float32(2)
		proRate<- pr
	}
	p := 384.7019274542
	a1 := -0.1576960777
	a2 := -0.5020923248
	a3 := -0.6671023113
	if t == 2 {
		pr = float32(float64(hWin[9]) * a1 + float64(hWin[8]) * a2 + float64(hWin[7]) * a3 + p)
        proRate<- pr
	}
}

func scale(t int, n int) {
	if t == 1 {
		cmd := exec.Command("./ScaleVmUp", strconv.Itoa(n))
		if _, err := cmd.Output(); err != nil {
			log.Println(err)
		}
		totalServer += n
		time.Sleep(time.Duration(60) * time.Second)
	}

	if t == 2 {
		cmd := exec.Command("./ScaleDockerUp", strconv.Itoa(n))
		if _, err := cmd.Output(); err != nil {
			log.Println(err)
		}
		time.Sleep(time.Duration(15) * time.Second)
		totalServer += n
	}

	if t == 3 {
		cmd := exec.Command("./ScaleDown", strconv.Itoa(1))
		if _, err := cmd.Output(); err != nil {
			log.Println(err)
		}
		totalServer -= 1
		time.Sleep(time.Duration(15) * time.Second)
	}
}

func scaler() {
	log.Println("Scaler is Running")
	for {
		select {
		case t := <- scaleType:
			{
				n := <- numServer
				scale(t, n)
			}
		}
	}

}

func UsageAndExit() {
	fmt.Fprint(os.Stderr, usage)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 3 {
		UsageAndExit()
	}

/*	var wg sync.WaitGroup
	wg.Add(4)
	wg.Wait()
*/

	totalServer = 1

	webAddress = os.Args[3]
	kupper = 0.5
	rateupper = 6000.0
	ttlupper = 1.0
	ttllower = 0.3

//	modelType  = make(chan int)
	curTTL = make(chan float32)
	curRate = make(chan float32)
	proRate = make(chan float32)
	scaleType = make(chan int)
	monRate = make(chan float32, 10)
	numServer = make(chan int)
	//kRate = make(chan float32)

	hWin = make([]float32, 0, 10)
	pWin = make([]float32, 0, 4)

	go modeler(1)
	go scaler()
	go alerter(15)
	go monitor()
}
