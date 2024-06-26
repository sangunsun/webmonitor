# webmonitor
这是一个网页内容监测器，可监测多个网页，如果网页内容变化就通过邮件进行报警

//TODO :如果存在未过滤动态参数的动态网页，且要强制进入监控模式；则不监控未过滤动态参数的动态网页  
//TODO :且把该网页url存入change.txt中，并发邮件通知该网页未进入监控。  
//TODO ：把日志由fmt.println改为log.Println以加上时间。  
### 修改记录

+ 2020.3.10
	加入直接发送网页差异内容到邮箱功能  
	加入强制进入监控功能（而不管程序启动的时候被监测网页是否正常）
	
	修改日志记录方法为一天两个日志文件（两个文件分别是‘日期old.html'和'日期now.html'两个文件）

### 使用说明：（写给二进制文件使用者看的，而不是写给源码阅读者看的）

 
+ #### 二、本目录下初始共有三个文件 webmonitor.exe、urls.txt、config.txt 及一个文件夹log  
  + 其中webmonitor为主程序，在shell下执行就可以；urls.txt为要监控的url地址列表，一个url一行，如果该url含有动态参数 在url后使用~符号后跟过滤用正则表达式过滤动态参数  
  + config.txt为网页发生变更时需要发送邮件的配置文件，其中第一行配置发送服务器信息，依次为smtp服务器地址，端口号,用户名,密码,各参数用逗号分隔  
  + 第二行开始为接收邮箱的地址列表，一个邮箱一行  
  + log程序产生的日志存放于log目录下，网页发生变化时的日志文件，一天一份（两个文件分别是‘日期old.html'和'日期now.html'两个文件）
+ #### 三、根据监控情况的不同可能会在log文件夹下生成如下文件  
  1. curl文件：如果进行监控准备时发现有网页前后两次读取的网页不一致，该网页的url地址会保存于这个文件  
  2. oldpages文件：如果进行监控准备时发现有网页前后两次读取的网页不一致，该网页第一次读取的的内容会保存于这个文件，如果有多个不一致的网页会一并保存  
  3. nowpages文件：如果进行监控准备时发现有网页前后两次读取的网页不一致，该网页第二次读取的的内容会保存于这个文件，如果有多个不一致的网页会一并保存  
  4. change文件：如果监控的时候发现有网页改变了，发生改变的网页的url 和发现网页改变的时间会记录在这个文件里，别外程序的启动时间也会存入该文件  
  5. old.html文件：如果监控中发现某个网页发生了变化，则该网页的URL地址、发现变化的时间、变化前的内容会写入该文件  
  6. now.html文件：如果监控中发现某个网页发生了变化，则该网页的URL地址、发现变化的时间、变化后的内容会写入该文件  

+ #### 四、如何使用  
1. 首先把需要监控的网页的url地址保存到urls.txt文件里，一个url一行。  
2. 配置邮箱配置文件config.txt  
3. 启动程序进行监控准备，如果存在有动态参数的网页，程序会退出  
4. 用文件比较工具查看oldpages.txt和nowpages.txt文件的不同之处，并写出过滤动态参数的正则表达式用~符号和url地址分隔保存到urls.txt文件的对应url后面  
5. 再次启动监控程序，如果过滤条件合适，程序在完成监控准备后会自动进入监控状态  
6. 如果在监控时发现有网页内容发生变更，程序会发邮件给config.txt文件配置的接收邮箱（可多个）  
7. 程序自动以新的网页做为参照，比较网页是否发生变更。
8. 程序在读网页错误时也会发邮件给邮箱，在发现发生故障的网页恢复正常时也会发邮件给邮箱。
9. 程序进行监控状态后会一直运行，要中断程序请使用ctrl+c 退出监控程序  

+ 注意：config.txt中的邮箱配置内容在程序运行的时候是可以动态修改的，修改的时候一次性修改正确再保存  
+ 如果监控时网页无法打开，会发邮件提醒，如果后面网页又可以打开，会再次发邮件提醒  
如果网页发生变动，会发邮件提醒，但同时会把变动后的网页做为基准网页继续监控，所以收到网页变动的邮件后要及时处理。  
