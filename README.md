为了家里小朋友,要下载一个视频,放在电视上看.
所以利用一点时间,写了这个.程序很简单,单纯只是下载,没优化代码,只是实现功能.

只有二个参数,下载失败的ts会重试下载四次:

-u 要下载的index.m3u8网址.(需要用google或360浏览器,进入开发者模式,按F12或ctrl+shift+c,在里面点network,再把网页刷新,搜索m3u8,就可找到index.m3u8文件的URL)

-c 并发数量,默认是15.可以不输入.

例如:

dm3u8.exe -u https://cn4.qxreader.com/hls/20200131/baeee825f6605d5ab28b954f07e24386/1580471232/index.m3u8

dm3u8.exe -c 30 -u https://www.mmicloud.com:65/20191204/I2jpA2LP/index.m3u8
