# 更新日志: 

1. 修复登录成功后, 无法正常使用的问题 ([#399](https://github.com/iikira/BaiduPCS-Go/issues/399));
2. 修复文件夹下载无法调整重试次数;
3. 使用 export 导出, 会输出导出错误的文件或目录了;
4. 更换失效的默认app_id, 对于不是全新安装此程序的用户, 仍然需要手动更改默认的app_id, ([#387](https://github.com/iikira/BaiduPCS-Go/issues/387))

```
BaiduPCS-Go config set -appid 266719
```


个人项目bug在所难免! 欢迎提 issue 和 pull request!!.

# 下载说明

## 解释 CPU架构

|amd|arm| mips| 说明 |
|-----|----------------|------------------|------------------|
|amd64, x64 |arm64   | mips64, mips64le |适用于64位CPU或操作系统的计算机|
|386, x86 |armv5, armv7  | mips, mipsle |适用于32位CPU或操作系统的计算机|

## 注意区别 `arm` 和 `amd`, 不要搞错了!!!!

## 下载

* PC/电脑: 
    请选择对应的系统 (windows, linux, darwin(苹果系统), freebsd), 对应的CPU架构 (一般情况下是 amd), 对应的CPU或操作系统位数 (详见上表), 下载.

* 移动设备: 
    请选择对应的系统(android, darwin(ios系统)), 对应的CPU架构 (一般情况下是 arm, 除了少数手机的CPU架构要选 amd, 例如联想K800, 联想K900等), 对应的CPU或操作系统位数  (详见上表), 下载.

## 注意

Android 5.0 以上的设备请不要下载使用linux版本的, 否则网络请求可能会出现问题.

相关的关键词, 均能在文件名中找到. 

文件格式均为zip压缩包格式, 切勿未解压程序就直接运行!! 程序解压之后才可以正常使用.
