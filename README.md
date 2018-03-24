# BaiduPCS-Go 百度网盘客户端

仿 Linux 文件处理命令的百度网盘命令行客户端.

This project was largely inspired by [GangZhuo/BaiduPCS](https://github.com/GangZhuo/BaiduPCS)

# 特色

多平台支持, 支持 Windows, macOS, linux, 移动设备等.

百度帐号多用户支持;

网盘内列出文件和目录, **支持通配符匹配路径**, [通配符_百度百科](https://baike.baidu.com/item/通配符);

下载网盘内文件, 支持网盘内目录 (文件夹) 下载, 支持多个文件或目录下载, 支持断点续传和高并发下载;

# 程序 编译/交叉编译 说明
参见 [编译/交叉编译帮助](https://github.com/iikira/BaiduPCS-Go/wiki/编译-交叉编译帮助) 

# 程序 下载/运行 说明

Go语言程序, 可直接下载使用, [点此查看发布页面 / 下载汇总](https://github.com/iikira/BaiduPCS-Go/releases).

如果程序运行时输出乱码, 请检查下终端的编码方式是否为 `UTF-8`.

使用本程序之前, 建议学习一些 linux 基础知识 和 基础命令.

如果未带任何参数运行程序, 程序将会进入独有的 console 模式, 可直接运行相关命令.

console 模式下, 光标所在行的前缀应为 `BaiduPCS-Go >`, 如果登录了百度帐号则格式为 `BaiduPCS-Go:<工作目录> <百度ID>$ `

程序会提供相关命令的使用说明.

## Windows

程序应在 命令提示符 (Command Prompt) 或 PowerShell 中运行, 在 mintty (例如: GitBash) 可能会有显示问题.

也可直接双击程序运行, 具体使用方法请参见 [命令列表及说明](#命令列表及说明) 和 [例子](#举一些例子).

## Linux / macOS

程序应在 终端 (Terminal) 运行.

具体使用方法请参见 [命令列表及说明](#命令列表及说明) 和 [例子](#举一些例子).

## Android / iOS

> Android / iOS 移动设备操作比较麻烦, 本人不建议在移动设备上使用本程序.

安卓, 建议使用软件 [Termux](https://termux.com) 或 [NeoTerm](https://github.com/NeoTerm/NeoTerm/releases) 或 终端模拟器, 以提供终端环境.

示例: [Android 运行本 BaiduPCS-Go 程序参考示例](https://github.com/iikira/BaiduPCS-Go/wiki/Android-运行本-BaiduPCS-Go-程序参考示例), 有兴趣的可以参考一下.

苹果iOS, 需要越狱, 在 Cydia 搜索下载并安装 MobileTerminal, 或者其他提供终端环境的软件.

具体使用方法请参见 [命令列表及说明](#命令列表及说明) 和 [例子](#举一些例子).

# 命令列表及说明

## 注意 ! ! !

命令的前缀 `BaiduPCS-Go` 为指向程序运行的全路径名 (ARGv 的第一个参数)

直接运行程序时, 未带任何其他参数, 则程序进入 console 模式, 运行以下命令时, 要把命令的前缀 `BaiduPCS-Go` 去掉!

console 模式已支持按tab键自动补全命令, 后续会添加更多的自动补全规则.

## 登录百度帐号

### 常规登录百度帐号

支持在线验证绑定的手机号或邮箱,
```
BaiduPCS-Go login
```

### 使用百度 BDUSS 来登录百度帐号

[关于 获取百度 BDUSS](https://github.com/iikira/BaiduPCS-Go/wiki/关于-获取百度-BDUSS)

```
BaiduPCS-Go login -bduss=<BDUSS>
```

#### 例子
```
BaiduPCS-Go login -bduss=1234567
```
```
BaiduPCS-Go login
请输入百度用户名(手机号/邮箱/用户名), 回车键提交 > 1234567
```

## 获取当前帐号, 和所有已登录的百度帐号
```
BaiduPCS-Go loglist
```

## 切换已登录的百度帐号
```
BaiduPCS-Go su <uid>
```
```
BaiduPCS-Go su

请输入要切换帐号的 index 值 > 
```

## 退出当前登录的百度帐号
```
BaiduPCS-Go logout
```

程序会进一步确认退出帐号, 防止误操作.

## 获取配额, 即获取网盘总空间, 和已使用空间
```
BaiduPCS-Go quota
```

## 切换工作目录
```
BaiduPCS-Go cd <目录>
```

### 切换工作目录后自动列出工作目录下的文件和目录
```
BaiduPCS-Go cd -l <目录>
```

#### 例子
```
# 切换 /我的资源 工作目录
BaiduPCS-Go cd /我的资源

# 切换 /我的资源 工作目录, 并自动列出 /我的资源 下的文件和目录
BaiduPCS-Go cd -l 我的资源

# 使用通配符
BaiduPCS-Go cd /我的*
```

## 输出当前所在目录
```
BaiduPCS-Go pwd
```

## 列出当前工作目录的文件和目录或指定目录
```
BaiduPCS-Go ls
```
```
BaiduPCS-Go ls <目录>
```

#### 例子
```
BaiduPCS-Go ls 我的资源

# 使用通配符
BaiduPCS-Go ls /我的*
```

## 获取单个文件/目录的元信息 (详细信息)
```
BaiduPCS-Go meta <文件/目录>
```
```
# 默认获取工作目录元信息
BaiduPCS-Go meta
```

#### 例子
```
BaiduPCS-Go meta 我的资源
BaiduPCS-Go meta /
```

## 下载文件, 网盘文件或目录的绝对路径或相对路径
```
BaiduPCS-Go download <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...
BaiduPCS-Go d <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...
```

### 可选参数
```
-test: 测试下载, 此操作不会保存文件到本地
-p <num>: 指定下载的最大并发量
```

支持多个文件或目录的下载.

下载的文件默认保存到 **程序所在目录** 的 download/ 目录, 支持设置指定目录, 重名的文件会自动跳过!

#### 例子
```
# 设置保存目录, 保存到 D:\Downloads
# 注意区别反斜杠 "\" 和 斜杠 "/" !!!
BaiduPCS-Go config set -savedir D:/Downloads

# 下载 /我的资源/1.mp4
BaiduPCS-Go d /我的资源/1.mp4

# 下载 /我的资源 整个目录!!
BaiduPCS-Go d /我的资源

# 下载网盘内的全部文件!!
BaiduPCS-Go d /
BaiduPCS-Go d *
```

## 上传文件
```
BaiduPCS-Go upload <本地文件或目录的路径1> <文件或目录2> <文件或目录3> ... <网盘的目标目录>
BaiduPCS-Go u <本地文件或目录的路径1> <文件或目录2> <文件或目录3> ... <网盘的目标目录>
```

#### 例子:
```
# 将本地的 C:\Users\Administrator\Desktop\1.mp4 上传到网盘 /视频 目录
# 注意区别反斜杠 "\" 和 斜杠 "/" !!!
BaiduPCS-Go upload C:/Users/Administrator/Desktop/1.mp4 /视频

# 将本地的 C:\Users\Administrator\Desktop\1.mp4 和 C:\Users\Administrator\Desktop\2.mp4 上传到网盘 /视频 目录
BaiduPCS-Go upload C:/Users/Administrator/Desktop/1.mp4 C:/Users/Administrator/Desktop/2.mp4 /视频

# 将本地的 C:\Users\Administrator\Desktop 整个目录上传到网盘 /视频 目录
BaiduPCS-Go upload C:/Users/Administrator/Desktop /视频
```

## 手动秒传文件
```
BaiduPCS-Go rapidupload -length=<文件的大小> -md5=<文件的 md5 值> -slicemd5=<文件前 256KB 切片的 md5 值> -crc32=<文件的 crc32 值 (可选)> <保存的网盘路径, 需包含文件名>
BaiduPCS-Go ru -length=<文件的大小> -md5=<文件的 md5 值> -slicemd5=<文件前 256KB 切片的 md5 值> -crc32=<文件的 crc32 值 (可选)> <保存的网盘路径, 需包含文件名>
```

注意: 使用此功能秒传文件, 前提是知道文件的大小, md5, 前256KB切片的 md5, crc32 (可选), 且百度网盘中存在一模一样的文件.

#### 例子:
```
# 如果秒传成功, 则保存到网盘路径 /test
BaiduPCS-Go rapidupload -length=56276137 -md5=fbe082d80e90f90f0fb1f94adbbcfa7f -slicemd5=38c6a75b0ec4499271d4ea38a667ab61 -crc32=314332359 /test
```

## 获取文件的秒传信息
```
BaiduPCS-Go sumfile <本地文件的路径>
BaiduPCS-Go sf <本地文件的路径>
```

获取文件的大小, md5, 前256KB切片的 md5, crc32, 可用于秒传文件.

#### 例子:
```
# 获取 C:\Users\Administrator\Desktop\1.mp4 的秒传信息
BaiduPCS-Go sumfile C:/Users/Administrator/Desktop/1.mp4
```

## 创建目录
```
BaiduPCS-Go mkdir <目录>
```

#### 例子
```
BaiduPCS-Go mkdir 123
```

## 删除 单个/多个 文件/目录
```
BaiduPCS-Go rm <网盘文件或目录的路径1> <文件或目录2> <文件或目录3> ...
```

注意: 删除多个文件和目录时, 请确保每一个文件和目录都存在, 否则删除操作会失败.

被删除的文件或目录可在网盘文件回收站找回.

#### 例子
```
# 删除 /我的资源/1.mp4
BaiduPCS-Go rm /我的资源/1.mp4

# 删除 /我的资源/1.mp4 和 /我的资源/2.mp4
BaiduPCS-Go rm /我的资源/1.mp4 /我的资源/2.mp4

# 删除 /我的资源 内的所有文件和目录, 但不删除该目录
BaiduPCS-Go rm /我的资源/*

# 删除 /我的资源 整个目录 !!
BaiduPCS-Go rm /我的资源
```

## 拷贝(复制) 单个/多个 文件/目录
```
BaiduPCS-Go cp <文件/目录> <目标 文件/目录>
BaiduPCS-Go cp <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>
```

注意: 拷贝(复制) 多个文件和目录时, 请确保每一个文件和目录都存在, 否则拷贝操作会失败.

#### 例子
```
# 将 /我的资源/1.mp4 复制到 根目录 /
BaiduPCS-Go cp /我的资源/1.mp4 /

# 将 /我的资源/1.mp4 和 /我的资源/2.mp4 复制到 根目录 /
BaiduPCS-Go cp /我的资源/1.mp4 /我的资源/2.mp4 /
```

## 移动/重命名 单个/多个 文件/目录
```
# 移动: 
BaiduPCS-Go mv <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>
# 重命名: 
BaiduPCS-Go mv <文件/目录> <重命名的文件/目录>
```

注意: 移动多个文件和目录时, 请确保每一个文件和目录都存在, 否则移动操作会失败.

#### 例子
```
# 将 /我的资源/1.mp4 移动到 根目录 /
BaiduPCS-Go mv /我的资源/1.mp4 /

# 将 /我的资源/1.mp4 重命名为 /我的资源/3.mp4
BaiduPCS-Go mv /我的资源/1.mp4 /我的资源/3.mp4
```

## 离线下载
```
BaiduPCS-Go offlinedl
BaiduPCS-Go clouddl
BaiduPCS-Go od
```

离线下载支持http/https/ftp/电驴/磁力链协议

### 添加离线下载任务
```
BaiduPCS-Go offlinedl add -path=<离线下载文件保存的路径> 资源地址1 地址2 ...
```

添加任务成功之后, 返回离线下载的任务ID.

### 精确查询离线下载任务
```
BaiduPCS-Go offlinedl query 任务ID1 任务ID2 ...
```

### 查询离线下载任务列表
```
BaiduPCS-Go offlinedl list
```

### 取消离线下载任务
```
BaiduPCS-Go offlinedl cancel 任务ID1 任务ID2 ...
```

### 删除离线下载任务
```
BaiduPCS-Go offlinedl delete 任务ID1 任务ID2 ...
```

#### 例子
```
# 将百度和腾讯主页, 离线下载到根目录 /
BaiduPCS-Go offlinedl add -path=/ http://baidu.com http://qq.com

# 添加磁力链接任务
BaiduPCS-Go offlinedl add magnet:?xt=urn:btih:xxx

# 查询任务ID为 12345 的离线下载任务状态
BaiduPCS-Go offlinedl query 

# 取消任务ID为 12345 的离线下载任务
BaiduPCS-Go offlinedl cancel 12345
```

## 显示和修改程序配置项
```
BaiduPCS-Go config
BaiduPCS-Go config set
```

#### 例子
```
# 显示所有可以设置的值
BaiduPCS-Go config -h
BaiduPCS-Go config set -h

# 设置下载文件的储存目录
BaiduPCS-Go config set -savedir D:/Downloads

# 设置下载最大并发量为 150
BaiduPCS-Go config set -max_parallel 150

# 组合设置, 
BaiduPCS-Go config set -max_parallel 150 -savedir D:/Downloads
```

# 举一些例子 

新手建议: **双击运行程序**, 进入 console 模式;

console 模式下, 光标所在行的前缀应为 `BaiduPCS-Go >`, 如果登录了百度帐号则格式为 `BaiduPCS-Go:<工作目录> <百度ID>$ `

以下例子的命令, 均为 console 模式下的命令

运行命令的正确操作: **输入命令, 按一下回车键 (键盘上的 Enter 键)**, 程序会接收到命令并输出结果

## 1. 查看程序使用说明

console 模式下, 运行命令 `help`

## 2. 登录百度帐号 (必做)

console 模式下, 运行命令 `login -h` (注意空格) 查看帮助

console 模式下, 运行命令 `login` 程序将会提示你输入百度用户名(手机号/邮箱/用户名)和密码, 必要时还可以在线验证绑定的手机号或邮箱

## 3. 切换网盘工作目录

console 模式下, 运行命令 `cd /我的资源` 将工作目录切换为 `/我的资源` (前提: 该目录存在于网盘)

目录支持通配符匹配, 所以你也可以这样: 运行命令 `cd /我的*` 或 `cd /我的??` 将工作目录切换为 `/我的资源`, 简化输入.

将工作目录切换为 `/我的资源` 成功后, 运行命令 `cd ..` 切换上级目录, 即将工作目录切换为 `/`

为什么要这样设计呢, 举个例子, 

假设 你要下载 `/我的资源` 内名为 `1.mp4` 和 `2.mp4` 两个文件, 而未切换工作目录, 你需要依次运行以下命令: 

```
d /我的资源/1.mp4
d /我的资源/2.mp4
```

而切换网盘工作目录之后, 依次运行以下命令: 

```
cd /我的资源
d 1.mp4
d 2.mp4
```

这样就达到了简化输入的目的

## 4. 网盘内列出文件和目录

console 模式下, 运行命令 `ls -h` (注意空格) 查看帮助

console 模式下, 运行命令 `ls` 来列出当前所在目录的文件和目录

console 模式下, 运行命令 `ls /我的资源` 来列出 `/我的资源` 内的文件和目录

console 模式下, 运行命令 `ls ..` 来列出当前所在目录的上级目录的文件和目录

## 5. 下载文件

说明: 下载的文件将会保存到 download/ 目录 (文件夹)

console 模式下, 运行命令 `d -h` (注意空格) 查看帮助

console 模式下, 运行命令 `d /我的资源/1.mp4` 来下载位于 `/我的资源/1.mp4` 的文件 `1.mp4` , 该操作等效于运行以下命令: 

```
cd /我的资源
d 1.mp4
```

现在已经支持目录 (文件夹) 下载, 所以, 运行以下命令, 会下载 `/我的资源` 内的所有文件 (违规文件除外): 

```
d /我的资源
```

参见 例6 设置下载最大并发数

## 6. 设置下载最大并发数

console 模式下, 运行命令 `config set -h` (注意空格) 查看设置帮助以及可供设置的值

console 模式下, 运行命令 `config set -max_parallel 250` 将下载最大并发数设置为 250

下载最大并发数建议值: 50~500, 太低下载速度提升不明显甚至速度会变为0, 太高可能会导致程序出错被操作系统结束掉.

## 7. 退出程序

运行命令 `quit` 或 `exit` 或 组合键 `Ctrl+C` 或 组合键 `Ctrl+D`

# 已知问题

## 个人项目bug在所难免! 欢迎提 issue 和 pull request!!

1. 下载进度到最后的时候, 下载速度会降低;

2. 下载速度不是很稳定, 有时会上下波动.

# 常见问题

参见 [常见问题](https://github.com/iikira/BaiduPCS-Go/wiki/%E5%B8%B8%E8%A7%81%E9%97%AE%E9%A2%98)

# TODO

1. 上传大文件
2. 自定义下载文件的保存位置