# 用于直接从docker hub中下载镜像包

我们经常会遇到需要离线安装docker包的情况

如果每次都要安装docker，然后再去docker hub下载镜像包，这样的话，就会很麻烦，而且还会很慢

所以，我们可以直接使用dget从docker hub中下载镜像包，然后再离线安装


## 直接下载链接

[windows x64版本](./bin/windows_amd64/dget.exe)
[linux amd64版本](./bin/linux_amd64/dget)
[linux arm版本](./bin/linux_arm/dget)
[Mac 传统版本](./bin/darwin_amd64/dget)
[Mac arm64版本](./bin/darwin_arm64/dget)

## 使用go安装dget

```bash
go install gitee.com/extrame/dget/cmd/dget@latest
```

## 使用方法

注意，本程序为命令行程序，需要使用命令行[cmd/powershell/bash等]打开

```bash
dget influxdb:1.8.3
```

总之，就是dget后面跟docker镜像名，然后就会自动下载到当前目录的tmp_xxx目录下，下载有缓存支持，如果一次出错了，直接再次执行就可以了

成功的话，会直接生成tar.gz包

## 关于从第三方registry下载

```
dget alibaba-cloud-linux-3-registry.cn-hangzhou.cr.aliyuncs.com/alinux3/alinux3:220901.1
```

形如上述调用方法，直接在包名称前面跟上服务器地址即可（v1.0.1)

## 选择架构

最近很多的包都推出了多架构，命令增加了选择架构的功能

使用参数-arch可以指定下载的架构，例如 linux/arm等，请使用/分隔系统和架构，例如

```bash
dget -arch linux/arm influxdb:1.8.3
```

## 设置代理

使用参数 -proxy 设置下载和获取时需要使用的代理

## 获取tag

如果你不知道要获取那个tag的软件，可以使用-tag参数获得软件的tag列表，由@joder提供

```bash
dget -tag influxdb:1.8.3
```

## 直接下载链接

[windows x64版本](./bin/windows_amd64/dget.exe)
[linux amd64版本](./bin/linux_amd64/dget)
[linux arm版本](./bin/linux_arm/dget)
[Mac 传统版本](./bin/darwin_amd64/dget)
[Mac arm64版本](./bin/darwin_arm64/dget)
