<p align="center">
  <img src="build/appicon.png" alt="OpenCecs Logo" width="120" />
</p>

<h1 align="center">OpenCecs - ARM 边缘计算设备管理客户端</h1>

<p align="center">
  <strong>一站式管理 ARM 边缘计算设备与云手机实例的跨平台桌面客户端</strong>
</p>

<p align="center">
  <a href="#功能特性">功能特性</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#编译构建">编译构建</a> •
  <a href="#项目结构">项目结构</a> •
  <a href="#使用指南">使用指南</a> •
  <a href="#贡献指南">贡献指南</a> •
  <a href="#许可证">许可证</a>
</p>

---

## 📖 项目简介

OpenCecs（Open Cloud Edge Computing System）是一个基于 [Wails v3](https://wails.io/) 构建的跨平台桌面应用程序，用于管理 ARM 边缘计算设备（V0 ~ V3 版本）。它提供了设备发现、云手机（Android 容器）管理、镜像管理、VPC 网络配置、实时投屏等丰富功能，旨在为 ARM 边缘计算场景提供直观、高效的管理工具。

## ✨ 功能特性

### 🔍 设备发现与管理
- **自动发现**：通过 UDP 广播自动发现局域网内的 ARM 边缘计算设备
- **多版本支持**：兼容 V0、V1、V2、V3 多代设备
- **实时心跳监测**：基于 TCP Ping 的设备在线状态监控，低延迟响应
- **设备信息展示**：CPU 温度、内存用量、存储空间、网络状态等实时监控

### 📱 云手机管理
- **实例全生命周期管理**：创建、开机、关机、重启、删除云手机实例
- **批量操作**：支持批量创建、重启、重置、删除等操作
- **镜像管理**：在线镜像库、本地镜像上传、镜像版本切换
- **实时截图**：后台轮询获取实例截图，支持网格/列表两种展示模式
- **手机型号管理**：支持多种 Android 版本和手机型号的机型切换

### 🖥️ 实时投屏
- **低延迟投屏**：通过 scrcpy 协议实现云手机画面实时投射
- **触控操作**：支持鼠标操作映射为触控事件
- **多实例投屏**：同时打开多个云手机的投屏窗口
- **窗口置顶**：支持投屏窗口始终置顶

### 🌐 VPC 网络管理
- **代理配置**：支持 Tun2socks 和 Singbox 两种 S5 代理模式
- **代理服务器管理**：添加、删除、测试代理连接
- **网络状态检测**：实时检测代理连接状态

### 📡 流媒体推送
- **RTMP 推流**：支持将云手机画面推送到直播平台
- **虚拟摄像头**：支持 Windows 虚拟摄像头功能
- **P2P 推流**：基于 P2P 协议的低延迟推流方案

### 🤖 AI 助手 & RPA
- **AI 对话助手**：集成大语言模型，提供智能操作建议
- **RPA 自动化**：支持通过 RPA SDK 实现设备自动化操作

### 🔄 自动更新
- **在线更新检查**：自动检测新版本
- **增量更新**：支持 UAC 提权更新，无需手动下载

### 🌍 国际化
- **多语言支持**：内置中文、英文界面切换

## 🛠️ 环境要求

| 依赖 | 最低版本 | 说明 |
|------|---------|------|
| **Go** | 1.22+ | 后端编译（推荐 1.25+） |
| **Node.js** | 16+ | 前端构建 |
| **npm** | 8+ | 前端包管理 |
| **Wails CLI** | v3 (alpha) | 应用框架 CLI |
| **Task** | 3.x | 任务运行器（可选） |
| **MinGW-w64** | - | Windows 下 CGO 编译需要（仅编译 RPA 模块时需要） |

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### 安装 Task（可选）

```bash
# Windows (通过 Scoop)
scoop install task

# macOS
brew install go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
```

## 🚀 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/your-org/opencecs.git
cd opencecs
```

### 2. 安装前端依赖

```bash
cd frontend
npm install
cd ..
```

### 3. 开发模式运行

```bash
# 方式一：使用 Wails CLI（推荐）
wails3 dev -config ./build/config.yml

# 方式二：使用 Task
task dev
```

> 开发模式下，前端使用 Vite 热更新、后端源码变动自动重编译。

### 4. 在浏览器中访问

开发模式启动后，Wails 会自动打开桌面窗口。你也可以在浏览器中访问 `http://localhost:34115` 调试前端页面。

## 📦 编译构建

### 标准构建（无 RPA）

```bash
# 方式一：使用 Wails CLI
wails3 build

# 方式二：使用 Task
task build
```

构建产物位于 `bin/` 目录下。

### 带 RPA 模块构建（需要 MinGW）

RPA 模块使用 CGO 进行编译，需要 MinGW-w64 工具链：

```bash
task build:rpa
```

> 此命令会先编译 `rpa/libmytrpc` 静态库，再以 `cgorpa` build tag 编译整个应用。

### 构建 Windows 安装包（NSIS）

```bash
task package
```

### 手动构建（不使用 Task）

```bash
# 1. 构建前端
cd frontend
npm run build
cd ..

# 2. 编译后端
go build -tags "production devtools" -trimpath -buildvcs=false -o bin/MYTV3.exe .
```

### 构建参数说明

| Build Tag | 说明 |
|-----------|------|
| `production` | 生产模式，禁用开发工具 |
| `devtools` | 启用浏览器开发工具 |
| `cgorpa` | 启用 RPA SDK（需要 CGO） |

### 跨平台 Docker 构建（在非 Windows 平台编译 Windows 版本）

```bash
# 先构建 Docker 镜像
task setup:docker

# 然后交叉编译
task build
```

## 📁 项目结构

```
opencecs/
├── main.go                    # 应用入口
├── app.go                     # 核心业务逻辑（设备管理、云机操作等）
├── app_windows.go             # Windows 平台特定代码
├── app_unix.go                # Unix 平台特定代码
├── device_heartbeat.go        # 设备心跳检测（TCP Ping）
├── android_poll.go            # Android 容器列表后台轮询
├── ai_service.go              # AI 助手服务
├── rtmp_service.go            # RTMP 流媒体服务
├── p2p_manager.go             # P2P 推流管理
├── batch_task_service.go      # 批量任务服务
├── updater_service.go         # 自动更新服务
├── rpa_service.go             # RPA 自动化服务
├── rpa_service_stub.go        # RPA 存根（无 CGO 时使用）
├── windows_pusher_api.go      # Windows 推流器 API
├── go.mod                     # Go 模块依赖
├── wails.json                 # Wails 项目配置
├── Taskfile.yml               # Task 任务定义
│
├── frontend/                  # 前端项目（Vue 3）
│   ├── package.json           # 前端依赖
│   ├── vite.config.js         # Vite 构建配置
│   ├── index.html             # HTML 入口
│   └── src/
│       ├── main.js            # Vue 入口
│       ├── App.vue            # 主组件
│       ├── i18n.js            # 国际化配置
│       ├── style.css          # 全局样式
│       ├── components/        # 组件目录
│       │   ├── HostManagement.vue         # 主机管理
│       │   ├── CloudManagement.vue        # 云机管理
│       │   ├── StreamManagement.vue       # 流媒体管理
│       │   ├── modelManagement.vue        # 型号管理
│       │   ├── instanceManagement.vue     # 实例管理
│       │   ├── backupManagement.vue       # 备份管理
│       │   ├── networkmanagement.vue      # 网络/VPC 管理
│       │   ├── opencecsManagement.vue     # OpenCecs 设备管理
│       │   ├── aiAssistant.vue            # AI 助手
│       │   ├── rpaAgent.vue               # RPA 代理
│       │   ├── BatchTaskManagement.vue    # 批量任务管理
│       │   ├── BatchImport.vue            # 批量导入
│       │   ├── UpdateDialog.vue           # 更新对话框
│       │   └── ...
│       ├── locales/           # 国际化语言文件
│       └── services/          # 服务层
│
├── build/                     # 构建配置
│   ├── config.yml             # Wails 构建配置
│   ├── windows/               # Windows 构建脚本
│   ├── darwin/                # macOS 构建脚本
│   └── linux/                 # Linux 构建脚本
│
├── dget/                      # 文件下载库（本地替换版）
├── rpa/                       # RPA SDK（C++ 静态库源码）
├── player/                    # 投屏播放器源码
├── player_dist/               # 投屏播放器分发包
└── updater/                   # 自动更新模块
```

## 📖 使用指南

### 添加设备

1. 启动应用后，点击左侧面板的 **「添加设备」** 按钮
2. 输入设备 IP 地址（支持 `IP:端口` 格式，V3 设备默认端口 8000）
3. 也可以点击 **「自动发现」** 按钮通过 UDP 广播扫描局域网设备

### 云手机管理

1. 在左侧设备列表中选择一台在线设备
2. 进入 **「云机管理」** 标签页
3. 点击空闲坑位上的 **「创建」** 按钮，选择镜像和机型创建云手机
4. 创建成功后可进行 **开机、关机、重启、删除、更新镜像** 等操作
5. 点击实例截图即可打开 **实时投屏** 窗口

### 批量操作

1. 在 **「主机管理」** 标签页选择设备
2. 勾选需要操作的实例
3. 使用顶部工具栏的批量操作按钮（批量重启、批量删除等）

### VPC 网络配置

1. 进入 **「网络管理」** 标签页
2. 添加代理服务器信息（地址、端口、用户名、密码）
3. 选择代理模式（Tun2socks / Singbox）
4. 为云手机实例配置代理

## ⚙️ 配置说明

### 应用配置

应用配置位于 `wails.json`：

```json
{
  "name": "edgeclient",
  "outputfilename": "edgeclient",
  "frontend:install": "npm install",
  "frontend:build": "npm run build"
}
```

### 版本配置

版本信息位于 `build/config.yml`：

```yaml
info:
  companyName: "MYT"
  productName: "MYTV3"
  version: "1.5.0"
```

### 调试模式

启动时传入 `--debug` 参数或设置环境变量 `APP_DEBUG=true` 可启用调试模式：

```bash
# 命令行参数
./edgeclient.exe --debug

# 环境变量
set APP_DEBUG=true
./edgeclient.exe
```

调试模式下将启用右键菜单和开发者工具。

## 🏗️ 技术栈

### 后端
- **Go 1.22+** — 高性能后端逻辑
- **[Wails v3](https://wails.io/)** — 桌面应用框架（Go + WebView）
- **Docker API** — 容器管理（V0-V2 设备）
- **自研 V3 API** — V3 设备通信协议

### 前端
- **[Vue 3](https://vuejs.org/)** + Composition API — UI 框架
- **[Element Plus](https://element-plus.org/)** — UI 组件库
- **[Vite](https://vitejs.dev/)** — 构建工具
- **[Axios](https://axios-http.com/)** — HTTP 客户端
- **[xterm.js](https://xtermjs.org/)** — 终端模拟器

### 通信协议
| 协议 | 用途 | 默认端口 |
|------|------|---------|
| UDP 广播 | 设备发现 | — |
| HTTP REST | 设备 API 通信 | V0-V2: 81, V3: 8000 |
| Wails IPC | 前后端通信 | — |
| Docker API | 容器管理 | 2375 |
| SSE | 实时进度推送 | — |
| RTMP | 流媒体推流 | — |

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

### 代码规范

- Go 代码遵循 [Effective Go](https://go.dev/doc/effective_go) 规范
- 前端代码遵循 Vue 3 官方风格指南
- 提交信息遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范

## ❓ 常见问题

### Q: 编译时报 `CGO_ENABLED` 相关错误？
**A:** RPA 模块需要 CGO 支持。如果你不需要 RPA 功能，使用标准构建即可（`wails3 build`），默认不启用 CGO。如果需要 RPA，请确保安装了 MinGW-w64 工具链。

### Q: 前端依赖安装失败？
**A:** 请确保 Node.js 版本 ≥ 16。可以尝试清除缓存后重新安装：
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Q: 设备发现不到设备？
**A:** 请确保：
1. 客户端与设备在同一局域网内
2. 防火墙未阻止 UDP 广播
3. 设备已开机且管理服务正常运行

### Q: 投屏功能无法使用？
**A:** 请确保 `player_dist/` 目录下包含投屏播放器的必要文件（`player.exe` 及相关 DLL）。

## 📄 许可证

本项目采用 [Apache License 2.0](LICENSE) 开源许可证。

## 📧 联系方式

- **作者**：Willzen
- **邮箱**：willzen@yeah.net
- **官网**：[https://www.moyunteng.com](https://www.moyunteng.com)

---

<p align="center">
  如果这个项目对你有帮助，请给一个 ⭐ Star 支持！
</p>
