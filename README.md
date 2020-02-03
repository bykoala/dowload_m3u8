为了家里小朋友,要下载一个视频,放在电视上看.所以利用一会时间,写了这个.程序很简单,单纯只是下载,其它什么都没处理.

只有二个参数:

-u 要下载的index.m3u8网址.(需要用google或360浏览器,进入开发者模式,按F12或ctrl+shift+c,在里面点network,再把网页刷新,搜索m3u8,就可找到index.m3u8文件的URL)

-c 并发数量,默认是15.可以不输入.

例如:
C:\Downloads>dm3u8.exe -u https://cn4.qxreader.com/hls/20200131/baeee825f6605d5ab28b954f07e24386/1580471232/index.m3u8
