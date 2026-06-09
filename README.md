# WinRDP-Changer
3389端口的修改，win11账号账号密码的修改
# 🛡️ WinRDP-Changer (Windows 运维小工具)

![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?style=flat-square&logo=go)
![Platform](https://img.shields.io/badge/Platform-Windows_11%20%7C%2010-blue?style=flat-square&logo=windows)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

这是一个使用 Go 语言编写的轻量级、无依赖的 Windows 系统高级运维工具。

本工具的核心功能是**一键全自动修改 Windows 远程桌面 (RDP) 的默认 3389 端口**。传统修改方式需要手动改动多处注册表并配置防火墙，极易出错导致无法远程连接；本工具实现了**修改注册表 -> 添加防火墙入站规则 -> 重启服务**的全自动无缝衔接。

除了端口修改，它还集成了让许多运维人员头疼的 **Windows 自动更新彻底禁用** 以及 **本地提权账户管理** 功能。

---

## ✨ 核心功能 (Features)

1. 🚪 **一键修改 RDP 端口 (告别 3389)**
   * 自动修改 `Wds` 和 `WinStations` 注册表核心键值。
   * 自动调用 `netsh` 在高级安全 Windows 防火墙中放行新端口。
   * 自动重启 `TermService` 远程桌面服务，无需重启电脑，即刻生效。
2. 🛑 **接管 Windows 自动更新**
   * **彻底禁用**：同时拦截 `wuauserv` (更新服务) 和 `UsoSvc` (更新协调程序)，并在组策略注册表写入锁死键 (`NoAutoUpdate=1`)。
   * **一键恢复**：随时解除锁定，恢复系统正常的自动更新下载功能。
3. 🔑 **本地账户密码管理**
   * 快速强制修改系统内任意本地用户的密码。
4. 👤 **创建全新管理员账户**
   * 一键创建隐藏或备用的本地账号，并自动将其加入最高权限的 `Administrators` 管理员组，方便独立远程登录。

---

## 🚀 快速开始 (Quick Start)

### 1. 获取源码并编译

由于程序调用了 Windows 底层 API 修改注册表和系统服务，推荐你在本地自行编译生成可执行文件，以确保安全。

确保你的电脑已安装 [Go 环境](https://go.dev/)，然后在终端执行以下命令：

```bash
# 1. 初始化 Go 模块
go mod init win-rdp-changer

# 2. 下载 Windows 系统库依赖
go get golang.org/x/sys/windows/registry

# 3. 编译生成 exe 可执行文件 (添加了 -ldflags 参数可减小体积并隐藏控制台黑框)
go build -ldflags="-s -w" -o sys_tool.exe main.go
