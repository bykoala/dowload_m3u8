package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DowloadPath = "./ts/" //下载保存路径.
	TryNum      = 3       //下载失败,再次重试几次
)

var (
	UrlInfo = urlInfo{} //url信息.
	TsUrls  = []ts{}    //ts文件的信息.
	TryTs   = []ts{}    //失败再次尝试下载
	TsChan  chan ts     //下载Ts文件的管道.
	Failed  chan ts
)

//url信息
type urlInfo struct {
	Url           string
	Path          string //Url的前缀,不包含后面的文件.
	Host          string //仅主域名.
	M3u8IndexFile string //M3u8的索引文件.
	tsNum         int    //M3u8的索引文件中总ts文件数.
}

//ts文件的信息
type ts struct {
	index      int    //ts文件索引号.
	tsUrl      string //ts文件的URL地址.
	suffix     string //文件后缀
	isDownload bool
}

//初始化所有信息,并将初始化数据存入管道
func initAll(url string) (err error) {
	//初始化基本信息
	err = initInfo(url)
	if err != nil {
		return err
	}
	//如果有下一层,则重新获取信息
	if len(TsUrls) == 1 {
		url = TsUrls[0].tsUrl
		err = initInfo(url)
		if err != nil {
			return err
		}
	}
	//初始化管道
	initchan()
	return nil
}

//初始化信息.
func initInfo(url string) (err error) {
	//get m3u8 URL message
	err = getUrlInfo(url)
	if err != nil {
		return err
	}
	//dowload M3U8 file
	err = getM3u8(url)
	if err != nil {
		return err
	}
	//get ts file from dowload M3U8 file
	err = getTsUrls(DowloadPath + UrlInfo.M3u8IndexFile)
	if err != nil {
		return err
	}
	return nil
}

//获取m3u8的URL的信息
func getUrlInfo(url string) (err error) {
	i := strings.Index(url, `http`)
	if i == -1 {
		return fmt.Errorf("m3u8 URL is invalid")
	}
	j := strings.LastIndex(url, `//`)
	if j == -1 {
		return fmt.Errorf("m3u8 URL is invalid")
	}
	temp := url[j+2:]
	host := url[:j+2]
	n := strings.Index(temp, `/`)
	if n == -1 {
		return fmt.Errorf("m3u8 URL is invalid")
	}
	temp = temp[:n]
	host = host + temp

	k := strings.LastIndex(url, `/`)
	if k == -1 {
		return fmt.Errorf("m3u8 URL is invalid")
	}
	UrlInfo = urlInfo{
		Url:           url,
		Path:          url[:k+1],
		Host:          host,
		M3u8IndexFile: url[k+1:],
	}
	return nil
}

//获取m3u8文件
func getM3u8(url string) (err error) {
	body, err := getUrl(url)
	if err != nil {
		return fmt.Errorf("m3u8 URL request failed:\n %w", err)
	}
	defer body.Close()

	err = getUrlInfo(url)
	if err != nil {
		return fmt.Errorf("Split Url: %s,\n %w", url, err)
	}

	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll err: %s,\n %w", url, err)
	}

	err = mkdir(DowloadPath)
	if err != nil {
		return fmt.Errorf("mkdir err: %s,\n %w", url, err)
	}

	err = ioutil.WriteFile(DowloadPath+UrlInfo.M3u8IndexFile, bytes, 0644)
	if err != nil {
		return fmt.Errorf("get M3u8 index file failed: %s,\n %w", url, err)
	}
	return nil
}

//获取ts的信息
func getTsUrls(fileName string) (err error) {
	var num int = 0
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		err = fmt.Errorf("open file(%s)failed:%w\n", fileName, err)
		return err
	}
	defer f.Close()
	fileScanner := bufio.NewScanner(f)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		line = strings.TrimSpace(line)
		// 以#或;开头视为注释,空行和注释不读取
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, ";") {
			continue
		}
		var tsUrl string
		var suffix string
		i := strings.LastIndex(line, ".")
		if i == -1 {
			suffix = ".ts"
			//continue
		} else {
			suffix = line[i:]
		}

		j := strings.Index(line, `/`)
		if j == -1 {
			tsUrl = UrlInfo.Path + line
		} else {
			tsUrl = UrlInfo.Host + line
		}

		if strings.Contains(line, "http") {
			tsUrl = line
		}
		if num == 0 && len(TsUrls) == 1 {
			TsUrls[0].index = num
			TsUrls[0].tsUrl = tsUrl
			TsUrls[0].suffix = suffix
			TsUrls[0].isDownload = false
		} else {
			ts := ts{
				index:      num,
				tsUrl:      tsUrl,
				suffix:     suffix,
				isDownload: false,
			}
			TsUrls = append(TsUrls, ts)
		}
		num++
	}
	UrlInfo.tsNum = num

	return nil
}

//获取url的数据
func getUrl(url string) (io.ReadCloser, error) {
	client := http.Client{
		Timeout: 120 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get error,\n %s", err)
	}
	return resp.Body, nil
}

//判断目录是否存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//创建目录，如果没有就创建。
func mkdir(dir string) (err error) {
	exist, err := pathExists(dir)
	if err != nil {
		return fmt.Errorf("get dir error!: %s", err)
	}
	if !exist {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("mkdir failed![%v]\n", err)
		}
	}
	return nil
}

//初始化管道,把需要下载的放入管道.
func initchan() {
	TsChan = make(chan ts, UrlInfo.tsNum)
	Failed = make(chan ts, UrlInfo.tsNum)
	for _, ts := range TsUrls {
		TsChan <- ts
	}
	close(TsChan)
}

//合并ts文件
func tsMerge() (err error) {
	rand.Seed(time.Now().UnixNano())
	MergeFile := "movie" + strconv.Itoa(rand.Intn(1000)) + ".ts"
	file, err := os.Create(MergeFile)
	if err != nil {
		return fmt.Errorf("create merge file failed：%s", err)
	}
	defer file.Close()
	writer := bufio.NewWriterSize(file, 32768)
	count := 0
	for _, ts := range TsUrls {
		tsPath := DowloadPath + strconv.Itoa(ts.index) + ts.suffix

		bytes, err := ioutil.ReadFile(tsPath)
		if err != nil {
			//return fmt.Errorf("Read merge file %s failed：%w",tsPath,err)
			continue
		}
		_, err = writer.Write(bytes)
		if err != nil {
			//return fmt.Errorf("Write merge file %s failed：%w",tsPath,err)
			continue
		}
		err = writer.Flush()
		if err != nil {
			continue
			//return fmt.Errorf("merge file failed：%s", err)
		}
		count++
	}

	err = os.RemoveAll(DowloadPath)
	if err != nil {
		return fmt.Errorf("delete temp dowload file failed：%s", err)
	}
	if count != UrlInfo.tsNum {
		return fmt.Errorf("[warning] %d Missing ts file download , dowload failed", UrlInfo.tsNum-count)
	}

	return nil
}

//下载Ts文件
func dowload() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("dowload panic", err)
			waitGroup.Done()
		}
	}()
	for ts := range TsChan {
		if ts.isDownload == true {
			continue
		}
		body, err := getUrl(ts.tsUrl)
		if err != nil {
			ts.isDownload = false
			Failed <- ts
			fmt.Println("getUrl error():", err)
			continue
		}
		defer body.Close()
		bytes, err := ioutil.ReadAll(body)
		if err != nil {
			ts.isDownload = false
			Failed <- ts
			fmt.Println("ioutil.ReadAll():", strconv.Itoa(ts.index)+ts.suffix, err)
			continue
		}
		Path := DowloadPath + strconv.Itoa(ts.index) + ts.suffix
		err = ioutil.WriteFile(Path, bytes, 0644)
		if err != nil {
			ts.isDownload = false
			Failed <- ts
			fmt.Println("ioutil.WriteFile():", err)
			continue
		}
		ts.isDownload = true
		fmt.Println("TS file download completed :", Path)
	}
	waitGroup.Done()
}

//尝试重新下载失败的
func tryFailed() {
	close(Failed)
	if len(Failed) < 1 {
		return
	}
	fmt.Println("try dowload Failed M3U8 url ts file.")
	for ts := range Failed {
		if ts.isDownload == true {
			continue
		}
		body, err := getUrl(ts.tsUrl)
		if err != nil {
			TryTs = append(TryTs, ts)
			fmt.Println("getUrl error():", err)
			continue
		}
		defer body.Close()
		bytes, err := ioutil.ReadAll(body)
		if err != nil {
			TryTs = append(TryTs, ts)
			fmt.Println("ioutil.ReadAll():", err)
			continue
		}
		Path := DowloadPath + strconv.Itoa(ts.index) + ts.suffix
		err = ioutil.WriteFile(Path, bytes, 0644)
		if err != nil {
			TryTs = append(TryTs, ts)
			fmt.Println("ioutil.WriteFile():", err)
			continue
		}
		ts.isDownload = true
		fmt.Println("dowload ts file finish:", Path)
	}
	//再次尝试
	tryFailed2()
}

//尝试重新下载失败的
func tryFailed2() {
	for i := 0; i < TryNum; i++ {
		if len(TryTs) < 1 {
			return
		}
		for _, ts := range TryTs {
			if ts.isDownload == true {
				continue
			}
			body, err := getUrl(ts.tsUrl)
			if err != nil {
				ts.isDownload = false
				fmt.Println("getUrl error():", err)
				continue
			}
			defer body.Close()
			bytes, err := ioutil.ReadAll(body)
			if err != nil {
				ts.isDownload = false
				fmt.Println("ioutil.ReadAll():", err)
				continue
			}
			Path := DowloadPath + strconv.Itoa(ts.index) + ts.suffix
			err = ioutil.WriteFile(Path, bytes, 0644)
			if err != nil {
				ts.isDownload = false
				fmt.Println("ioutil.WriteFile():", err)
				continue
			}
			ts.isDownload = true
			fmt.Println("dowload ts file finish:", Path)
		}
	}
}
