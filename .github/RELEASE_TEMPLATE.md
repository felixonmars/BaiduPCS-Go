# 更新日志: 

1. 新增上传、下载设置限速
2. 优化Windows下载大文件时的处理性能, [点此查看详情和设置方法](https://github.com/iikira/BaiduPCS-Go/wiki/Windows%E5%A6%82%E4%BD%95%E5%BC%80%E5%90%AF%E4%BC%98%E5%8C%96%E4%B8%8B%E8%BD%BD%E5%A4%A7%E6%96%87%E4%BB%B6%E6%97%B6%E7%9A%84%E5%A4%84%E7%90%86%E6%80%A7%E8%83%BD)
3. 更新下载器
4. 多处bug修复

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

* Android: 
    选择对应的CPU架构 (一般情况下是 arm, 除了少数手机的CPU架构要选 amd, 例如联想K800, 联想K900等), 对应的CPU或操作系统位数  (详见上表), 下载.

* iOS:
    无需选择CPU架构，选择 darwin-ios 下载解压后即可使用. 注意: armv7s架构的设备 (iPhone 5, iPhone 5c, iPad 4) 或 iOS 系统版本低于5.0, 可能无法正常运行.

## 注意

Android 5.0 以上的设备请不要下载使用linux版本的, 否则网络请求可能会出现问题.

相关的关键词, 均能在文件名中找到. 

文件格式均为zip压缩包格式, 切勿未解压程序就直接运行!! 程序解压之后才可以正常使用.
