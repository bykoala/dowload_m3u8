package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
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
	start := time.Now()
	flag.Parse()
	//url := "https://cn4.qxreader.com/hls/20200131/baeee825f6605d5ab28b954f07e24386/1580471232/index.m3u8"
	//url:="https://www.nmgxwhz.com:65/20200109/hvg9p8nT/index.m3u8"
	if url == "" {
		fmt.Println(`The parameter (u) "M3U8 url" must be entered.`)
		return
	}
	//一.初始化需要下载的信息
	fmt.Println("Start initialization M3U8 url index. Please wait.")
	err := initAll(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	//二.开始并发下载ts片段文件,默认设为15,可以命令行加-c调整.
	fmt.Println("Start dowload M3U8 url ts file.")
	for i := 0; i < maxGo; i++ {
		waitGroup.Add(1)
		go dowload()
	}
	waitGroup.Wait()

	//三.尝试重新下载失败的,默认重试下载3次.
	tryFailed()

	//四.将ts片段文件合并为一个文件
	fmt.Println("\nStart Merge ts file, Please wait.")
	err = tsMerge()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("\nDownload completed.")
	cost := time.Since(start)
	fmt.Printf("Total download time =[%s]\n", cost)
}
