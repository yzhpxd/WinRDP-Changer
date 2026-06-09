package main

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

func main() {
	for {
		fmt.Println("\n======================================")
		fmt.Println("    Windows 11 系统运维工具 (最终版)   ")
		fmt.Println("======================================")
		fmt.Println("1. 修改远程桌面(RDP)端口")
		fmt.Println("2. 禁用 Windows 11 自动更新")
		fmt.Println("3. 启用 Windows 11 自动更新")
		fmt.Println("4. 修改本地账户密码")
		fmt.Println("5. 创建新的管理员账户")
		fmt.Println("6. 退出程序")
		fmt.Print("请选择操作 (1-6): ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			setRdpPort()
		case 2:
			setWinUpdate(false)
		case 3:
			setWinUpdate(true)
		case 4:
			changeUserPassword()
		case 5:
			createAdminAccount()
		case 6:
			fmt.Println("谢谢使用，再见！")
			pause()
			return
		default:
			fmt.Println("无效选择，请重新输入。")
		}
	}
}

// 1. 设置远程桌面端口
func setRdpPort() {
	var port int
	fmt.Print("请输入新的远程桌面端口号 (1024-65535): ")
	fmt.Scanln(&port)

	if port < 1024 || port > 65535 {
		fmt.Println("【失败】端口号范围不合法！")
		return
	}

	paths := []string{
		`SYSTEM\CurrentControlSet\Control\Terminal Server\Wds\rdpwd\Tds\tcp`,
		`SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp`,
	}

	for _, path := range paths {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.SET_VALUE)
		if err != nil {
			fmt.Printf("【失败】无法打开注册表路径: %s, 错误: %v\n", path, err)
			return
		}
		defer key.Close()

		err = key.SetDWordValue("PortNumber", uint32(port))
		if err != nil {
			fmt.Printf("【失败】写入注册表失败: %v\n", err)
			return
		}
	}

	fmt.Printf("【成功】注册表端口已修改为: %d\n", port)
	fmt.Println("正在尝试自动添加防火墙入站规则...")
	
	ruleName := fmt.Sprintf("RDP_Custom_Port_%d", port)
	cmdStr := fmt.Sprintf("netsh advfirewall firewall add rule name=\"%s\" dir=in action=allow protocol=TCP localport=%d", ruleName, port)
	if runCmd(cmdStr) {
		fmt.Println("【成功】防火墙规则已添加！")
	}

	fmt.Println("正在重启远程桌面服务以使配置生效...")
	if runCmd("net stop TermService /y") && runCmd("net start TermService") {
		fmt.Println("【成功】远程桌面服务已重启！新端口已生效。")
	} else {
		fmt.Println("【提示】服务重启失败，请手动重启电脑使端口生效。")
	}
}

// 2 & 3. 开启或关闭 Windows 更新
func setWinUpdate(enable bool) {
	var stateCmd, startType string
	if enable {
		stateCmd = "start"
		startType = "demand"
		fmt.Println("正在开启 Windows 更新服务...")
	} else {
		stateCmd = "stop"
		startType = "disabled"
		fmt.Println("正在停止并禁用 Windows 更新服务...")
	}

	runCmd(fmt.Sprintf("net %s wuauserv", stateCmd))
	runCmd(fmt.Sprintf("sc config wuauserv start=%s", startType))
	runCmd(fmt.Sprintf("net %s UsoSvc", stateCmd))
	runCmd(fmt.Sprintf("sc config UsoSvc start=%s", startType))

	regPath := `SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate\AU`
	if !enable {
		key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, regPath, registry.SET_VALUE)
		if err == nil {
			_ = key.SetDWordValue("NoAutoUpdate", 1)
			key.Close()
		}
		fmt.Println("【成功】Windows 11 自动更新已彻底禁用！")
	} else {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.SET_VALUE)
		if err == nil {
			_ = key.DeleteValue("NoAutoUpdate")
			key.Close()
		}
		fmt.Println("【成功】Windows 11 自动更新已恢复正常！")
	}
}

// 4. 修改本地账户密码
func changeUserPassword() {
	var username, password string
	fmt.Print("请输入要修改密码的用户名 (例如 Administrator): ")
	fmt.Scanln(&username)
	if username == "" {
		fmt.Println("【失败】用户名不能为空。")
		return
	}

	fmt.Print("请输入该账户的新密码: ")
	fmt.Scanln(&password)
	if password == "" {
		fmt.Println("【失败】密码不能为空。")
		return
	}

	cmdStr := fmt.Sprintf("net user %s %s", username, password)
	if runCmd(cmdStr) {
		fmt.Printf("【成功】账户 [%s] 的密码已成功修改！\n", username)
	} else {
		fmt.Println("【失败】修改密码失败，请检查用户名是否存在或密码复杂度是否满足系统要求。")
	}
}

// 5. 创建新的管理员账户
func createAdminAccount() {
	var username, password string
	fmt.Print("请输入新账户的用户名: ")
	fmt.Scanln(&username)
	if username == "" {
		fmt.Println("【失败】用户名不能为空。")
		return
	}

	fmt.Print("请输入新账户的密码: ")
	fmt.Scanln(&password)
	if password == "" {
		fmt.Println("【失败】密码不能为空。")
		return
	}

	addUserCmd := fmt.Sprintf("net user %s %s /add", username, password)
	if !runCmd(addUserCmd) {
		fmt.Println("【失败】创建用户失败，可能该用户名已存在，或密码不符合Windows复杂度策略。")
		return
	}

	addGrpCmd := fmt.Sprintf("net localgroup administrators %s /add", username)
	if runCmd(addGrpCmd) {
		fmt.Printf("【成功】新管理员账户 [%s] 已成功创建并赋予管理员权限！\n", username)
	} else {
		fmt.Printf("【提示】用户 [%s] 已创建，但加入管理员组失败，请手动检查。\n", username)
	}
}

// 辅助函数：执行系统命令
func runCmd(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	head := parts[0]
	args := parts[1:]

	cmd := exec.Command(head, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

// 辅助函数：暂停屏幕显示
func pause() {
	fmt.Print("\n按回车键退出...")
	fmt.Scanln()
}