package main

import (
	"flag"
	"fmt"
	"sync"
)

var (
	url       string
	maxGo     int = 15
	waitGroup sync.WaitGroup
)

func init() {
	flag.StringVar(&url, "u", "", "M3U8 url index file")
	flag.IntVar(&maxGo, "c", 15, "maximum number of goroutine")
}
func main() {
	flag.Parse()

	//url := "https://cn4.qxreader.com/hls/20200131/baeee825f6605d5ab28b954f07e24386/1580471232/index.m3u8"
	//url:="https://www.mmicloud.com:65/20191204/I2jpA2LP/index.m3u8"
	if url == "" {
		fmt.Println(`The parameter (u) "M3U8 url" must be entered.`)
		return
	}

	fmt.Println("Start initialization M3U8 url index.")
	err := initAll(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Start dowload M3U8 url ts file.")
	for i := 0; i < maxGo; i++ {
		waitGroup.Add(1)
		go dowload()
	}

	waitGroup.Wait()

	//尝试重新下载失败的
	tryFailed()

	//合并文件
	fmt.Println("Start Merge  ts file, Please wait.")
	err = tsMerge()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Merge  ts file finish, Download completed.")
}
