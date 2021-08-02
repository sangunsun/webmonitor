
package main

import (
	"flag"
	"fmt"
	"io"

	"crypto/md5"
	"encoding/hex"
	//"net"
	"bufio"
	"crypto/tls"
	"errors"
	"github.com/sanmin33/diff"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var interval = flag.Int("interval", 10, "两次读取网页监控的间隔时间")
var fileName = flag.String("file", "urls.txt", "存放要扫描的ip列表文件名,每个ip一行，不留空行")
var help = flag.String("h", "", "help")
var forceMon = flag.String("f", "n", "是否强制进行监控状态")

type PagesStruct struct {
	sync.Mutex
	oldPages map[string]string
	oldMM    map[string]map[string]string //用于存放网页二级src资源的url及内容
	nowPages map[string]string
	nowMM    map[string]map[string]string //用于存放网页二级src资源的url及内容
	reString map[string]string
	isFalse  map[string]int
}
type locks struct {
	configLock sync.Mutex
	fileWLock  sync.Mutex
}

var myLocks locks
var pages PagesStruct

func main() {
	//1、读url列表文件并解析到字符串切片里
	//开URL数量个协程monAURL()对各URL进行监测
	fmt.Println("网页监测软件，对比两次读取的网页是否一致，自由软件，随意使用。")
	fmt.Println("")

	flag.Parse()
	pages.oldPages = make(map[string]string)
	pages.oldMM = make(map[string]map[string]string)

	pages.nowPages = make(map[string]string)
	pages.nowMM = make(map[string]map[string]string)

	pages.reString = make(map[string]string)
	pages.isFalse = make(map[string]int)
	writeToFile("./log/change.txt", fmt.Sprint(time.Now().Format("2006/1/2 15:04:05"))+" 监控程序启动。。。。\n")
	go func() {
		log.Println(http.ListenAndServe(":6660", nil))
	}()
	FileRead(*fileName)
	for runtime.NumGoroutine() > 1 {
		time.Sleep(time.Second * 3)
	}

}

func FileRead(name string) {
	fileObj, err := os.Open(name)
	if err != nil {
		fmt.Println("打开url列表文件失败，请检查文件名是否正确，默认文件名为urls.txt")
		return
	}
	defer fileObj.Close()
	//在定义空的byte列表时尽量大一些，否则这种方式读取内容可能造成文件读取不完整

	buf, err := ioutil.ReadAll(fileObj)
	if err != nil {
		fmt.Println("读取ip列表失败！")
		return
	}

	result := string(buf)
	//fmt.Println("原始文件内容:", result)

	//对ip列表文件的原始内容进行处理，去除多余空格 ，统一换行符
	result = strings.TrimSpace(result)
	result = strings.Replace(result, "\r\n", "\n", -1)

	urls := strings.Split(result, "\n")
	//fmt.Println("ip地址列表:", urls)

	//取出和url同一行的正则表达式对以后的url内容进行正则表达则过滤
	//原来urls中同时保存了url和该url对应的正则表达式过滤字符，把它拆分成两部分，分别保存到urls和pages.re中
	for i := 0; i < len(urls); i++ {
		urls[i] = strings.TrimSpace(urls[i])
		if nn := strings.Index(urls[i], "~"); nn != -1 {
			lurl := urls[i]
			url := strings.TrimSpace(lurl[0:nn])
			re := strings.TrimSpace(lurl[nn+1 : len(lurl)])
			pages.reString[url] = re
			urls[i] = url
		}

		fmt.Println(urls[i])
	}

	//初始化网页打开失败次数为0
	for i := 0; i < len(urls); i++ {
		pages.isFalse[urls[i]] = 0
	}

	// 间隔两次分别读每个url的信息，如果不一致，把两次数据不一致的url打印出来
	fmt.Println("正在第一次读取被监控的网页内容。。。")
	for i := 0; i < len(urls); i++ {
		fmt.Println("正在读网址：", urls[i])
		pages.oldPages[urls[i]], _ = readAURL(urls[i])

		//如果对应url存在正则表达式过滤串则过滤后保存
		if pages.reString[urls[i]] != "" {
			re, _ := regexp.Compile(pages.reString[urls[i]])
			pages.oldPages[urls[i]] = re.ReplaceAllString(pages.oldPages[urls[i]], "")
		}
	}

	time.Sleep(time.Second * 5)
	fmt.Println("正在第二次读取被监控的网页内容。。。")
	var haveDiffPage = false
	for i := 0; i < len(urls); i++ {
		fmt.Println("正在读网址：", urls[i])
		pages.nowPages[urls[i]], _ = readAURL(urls[i])
		//如果对应url存在正则表达式过滤串则过滤后保存
		if pages.reString[urls[i]] != "" {
			re, _ := regexp.Compile(pages.reString[urls[i]])
			pages.nowPages[urls[i]] = re.ReplaceAllString(pages.nowPages[urls[i]], "")
		}

		if pages.nowPages[urls[i]] != pages.oldPages[urls[i]] {
			fmt.Println(urls[i], "\n", "网页有动态变化参数,请在urls文件的url后用~做为分隔符加入过滤动态内容的正则表达式")

			fmt.Println("两次网页的内容分别保存于同目录下oldpages.html和nowpages.html文件中，请用文本比较工具查看差异部分\n")
			writeToFile("./log/curl.txt", urls[i]+"\n")
			writeToFile("./log/oldpages.html", urls[i]+"'\n"+pages.oldPages[urls[i]]+"\n")
			writeToFile("./log/nowpages.html", urls[i]+"'\n"+pages.nowPages[urls[i]]+"\n")
			haveDiffPage = true
		}
	}
	if haveDiffPage && *forceMon == "n" {
		fmt.Println("..........")
		fmt.Println("因存在有动态参数网页未处理，程序退出，具体网页差异见oldpages.html和nowpages.html，请处理后重新运行程序")
		fmt.Println("..........")
		os.Exit(0)
	}
	fmt.Println("开始监控网页。。。。")
	for i := 0; i < len(urls); i++ {
		go monAPage(urls[i])
	}

}

//readAURL根据传入的url读一个网页的内容，如果读不成功，返回空串和错误。
func readAURL(url string) (string, error) {
	var err error
	var client *http.Client
	if strings.Index(url, "https") == 0 {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {

		client = &http.Client{}
	}
	//生成请求
	reqest, err := http.NewRequest("GET", url, nil)

	//增加header选项,前两个参数是强制刷新不缓存网页数据
	reqest.Header.Add("Pragma", "no-cache")
	reqest.Header.Add("Cache-Control", "no-cache")
	reqest.Header.Add("User-Agent", "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0)")

	if err != nil {
		fmt.Println("打开网页失败", url, err)
		return "", err
	}

	//处理返回结果
	reqest.Close = true
	response, err := client.Do(reqest)

	if err != nil {
		fmt.Println("获取网页内容失败", url, err)
		return "", err
	}

	defer response.Body.Close()
	//TODO 处理网页返回的状态码
	if response.StatusCode != 200 && response.StatusCode != 302 {
		return "", errors.New("读网页错误，错误码：" + strconv.Itoa(response.StatusCode))
	}

	var page string
	pageByte, err := ioutil.ReadAll(response.Body)
	page = string(pageByte)
	if err != nil {
		return "", errors.New("读网页内容错误:" + err.Error())
	}
	if page == "" {
		return page, errors.New("读取到空页面，这是不正常的!!!")
	}
	return page, nil
}

func monAPage(url string) {
	var page string
	var err error

	defer func() {
		fmt.Println("程序异常退出了。。。", time.Now())
	}()

	for {
		fmt.Println("开始读:", url)
		page, err = readAURL(url)
		fmt.Println("完成读:", url, err)
		fmt.Println(runtime.NumGoroutine())
		if err != nil {
			// 发送网页无法打开的错误给预留邮箱。

			pages.isFalse[url]++
			if pages.isFalse[url] == 4 {
				fmt.Println("异常，无法打开网页", url, fmt.Sprint(err), time.Now())
				err = SendMail("读网页错误", url)

				_, err = writeToFile("./log/change.txt", fmt.Sprint(time.Now().Format("2006/1/2 15:04:05"))+" "+url+" 无法打开网页 "+fmt.Sprint(err)+"\n")
				if err != nil {
					fmt.Println("写入change.txt文件失败")
				}

			}

			time.Sleep(time.Second * time.Duration(*interval))
			continue
		}
		if pages.reString[url] != "" {
			re, _ := regexp.Compile(pages.reString[url])
			page = re.ReplaceAllString(page, "")
		}

		pages.Lock()
		pages.nowPages[url] = page
		pages.Unlock()

		if pages.isFalse[url] > 3 {
			fmt.Println("网页已恢复", url)

			_, err = writeToFile("./log/change.txt", fmt.Sprint(time.Now().Format("2006/1/2 15:04:05"))+" "+url+" 网页已恢复\n")
			if err != nil {
				fmt.Println("写入change.txt文件失败")
			}
			err = SendMail("网页已恢复", url)

		}
		//只要能正常读网页了，失败次数就要清零
		pages.isFalse[url] = 0

		if md5v1(pages.nowPages[url]) != md5v1(pages.oldPages[url]) {
			diffstr := diffTxt2(pages.oldPages[url], pages.nowPages[url])
			diffstr = "<pre>" + diffstr + "</pre>"
			//diffstr = strings.Replace(diffstr, "<", "&lt", -1)
			//diffstr = strings.Replace(diffstr, ">", "&gt", -1)
			fmt.Println(url, "\n", "warnning 网页已变动")
			fmt.Println(diffstr)
			fmt.Println("两次网页的内容分别保存于同目录下urlold.html和urlnow.html文件中，请用文本比较工具查看差异部分")

			preFileName := time.Now().Format("2006-01-02")

			_, err = writeToFile("./log/change.txt", fmt.Sprint(time.Now().Format("2006/1/2 15:04:05"))+" "+url+" 内容已变动\n")
			if err != nil {
				fmt.Println("写入change.txt文件失败")
			}

			oldFileName := "./log/" + preFileName + "old.html"
			_, err = writeToFile(oldFileName, url+"\n"+fmt.Sprintln(time.Now().Format("2006/1/2 15:04:05"))+pages.oldPages[url]+"\n")
			if err != nil {
				fmt.Println("写入url_old.html文件失败")
			}

			nowFileName := "./log/" + preFileName + "now.html"
			_, err = writeToFile(nowFileName, url+"\n"+fmt.Sprintln(time.Now().Format("2006/1/2 15:04:05"))+pages.nowPages[url]+"\n")
			if err != nil {
				fmt.Println("写入url_now.html文件失败")
			}

			// 发送网页已变动信息给预留邮箱
			fmt.Println("正在发送报警邮件....")

			err = SendMail("warnning 网页内容变动", url+"\n"+diffstr)

			//网页变动，并且发送邮件成功后，把新网页做为对比网页保存,这样可以防止不断的重复发送邮件。
			if err == nil {
				pages.oldPages[url] = pages.nowPages[url]
			}
		}

		time.Sleep(time.Second * time.Duration(*interval))
	}
}
func difftxt(str1 string, str2 string) string {
	/*	dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(str1, str2, true)
		out := dmp.DiffPrettyText(diffs)
		return string(out)
	*/
	out := diff.LineDiff(str1, str2)
	return out
}
func diffTxt2(str1 string, str2 string) string {
	fileName1 := md5v1(str1)
	fileName2 := md5v1(str2)
	writeToNewFile(fileName1, str1)
	writeToNewFile(fileName2, str2)
	strCmd := "diff " + fileName1 + " " + fileName2
	myCmd := exec.Command("bash", "-c", strCmd)
	byteDiff, err := myCmd.Output()

	os.Remove(fileName1)
	os.Remove(fileName2)
	fmt.Println(string(byteDiff), err)
	return string(byteDiff)

}
func SendMail(subject string, body string) error {

	var smptServer string
	var port string
	var user string
	var pass string

	//	myLocks.configLock.Lock()
	fileObj, err := os.Open("config.txt")

	if err != nil {
		fmt.Println("打开邮件配置文件失败确认已在config.txt文件中配置了邮箱信息")
		return nil
	}
	defer func() {
		if fileObj != nil {
			fileObj.Close()
		}
	}()
	//此处使用按行读取方式读email取配置文件中的发送邮箱配置情况
	rd := bufio.NewReader(fileObj)
	line, err := rd.ReadString('\n')
	if err == nil {
		emailLine := strings.Split(line, ",")
		if len(emailLine) != 4 {
			fmt.Println("邮件发送配置错误")
			return errors.New("邮件发送配置错误")
		}
		smptServer = strings.TrimSpace(emailLine[0])
		port = strings.TrimSpace(emailLine[1])
		user = strings.TrimSpace(emailLine[2])
		pass = strings.TrimSpace(emailLine[3])
	} else {
		fmt.Println("读邮件配置错误 ")
		return errors.New("读邮件配置错误")
	}

	fmt.Println("发送邮箱配置", smptServer, port, user)

	var toEmailAddress []string

	line = ""
	err = nil
	line, err = rd.ReadString('\n')

	for line != "" {

		//此处要对line进行去换行符号的处理
		line = strings.Replace(line, "\r\n", "", -1)
		line = strings.Replace(line, "\n", "", -1)
		line = strings.TrimSpace(line)
		if len(line) > 5 {
			toEmailAddress = append(toEmailAddress, line)
		}
		line, err = rd.ReadString('\n')
	}
	if fileObj != nil {
		fileObj.Close()
	}

	fmt.Println("目标邮件列表:", toEmailAddress)
	intPort, _ := strconv.Atoi(port) //转换端口类型为int

	m := gomail.NewMessage()

	//m.SetHeader("From", "网络安全报警"+"<"+user+">") //这种方式可以添加别名，
	m.SetHeader("From", user)
	m.SetHeader("To", toEmailAddress...) //发送给多个用户
	m.SetHeader("Subject", subject)      //设置邮件主题
	m.SetBody("text/plain", body)        //设置邮件正文

	d := gomail.NewDialer(smptServer, intPort, user, pass)
	//fmt.Println("发邮件", m, d)
	err = d.DialAndSend(m)

	fmt.Println("邮件发送结果:", time.Now(), err, "\n")
	//	myLocks.configLock.Unlock()
	return err

}
func writeToNewFile(wFileName string, conext string) (n int, err error) {
	var err1 error
	var f *os.File

	if checkFileIsExist(wFileName) { //如果文件存在
		err1 = os.Remove(wFileName)
		if err1 != nil {
			fmt.Println(err1)
			return -1, err1
		}

	}
	f, err1 = os.Create(wFileName) //创建文件
	if err1 != nil {
		fmt.Println(err1)
		return -1, err1
	}

	n, err1 = io.WriteString(f, conext) //写入文件(字符串)
	if err1 != nil {
		fmt.Println(err1)
		return -1, err1
	}
	f.Close()

	return n, err1
}
func writeToFile(wFileName string, conext string) (n int, err error) {
	var err1 error
	var f *os.File

	myLocks.fileWLock.Lock()
	fmt.Println(wFileName)
	if checkFileIsExist(wFileName) { //如果文件存在
		f, err1 = os.OpenFile(wFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend) //打开文件
		if err1 != nil {
			fmt.Println(err1)
		}

	} else {
		f, err1 = os.Create(wFileName) //创建文件
		if err1 != nil {
			fmt.Println(err1)
		}
	}

	n, err1 = io.WriteString(f, conext) //写入文件(字符串)
	if err1 != nil {
		fmt.Println(err1)
	}
	f.Close()
	myLocks.fileWLock.Unlock()

	return n, err1
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
func md5v1(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))

}
