package modules

import (
	"fmt"
	"runtime"
	"slices"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"

	"main/config"
	"main/config/helpers"
	"main/database"
)

func stats(msg *telegram.NewMessage) error {
	var text string
	msg.Delete()
	if !slices.Contains(config.OwnerId, msg.SenderID()) {
		msg.Respond("You are not authorised to use this command.")
		return telegram.EndGroup

	}
	if chats, err := database.GetServedChats(); err == nil {
		text += fmt.Sprintf("💬 <b>Total Chats:</b> %d\n", len(chats))
	}
	if users, err := database.GetServedUsers(); err == nil {
		text += fmt.Sprintf("👤 <b>Total Users:</b> %d\n", len(users))
	}

	// Bot uptime
	uptime := time.Since(config.StartTime)
	uptimeStr := helpers.FormatUptime(uptime)
	text += fmt.Sprintf("⏱️ <b>Bot Uptime:</b> %s\n", uptimeStr)

	// System uptime
	if upSecs, err := host.Uptime(); err == nil {
		sysUptime := time.Duration(upSecs) * time.Second
		text += fmt.Sprintf("🖥️ <b>System Uptime:</b> %s\n", helpers.FormatUptime(sysUptime))
	}

	text += "\n🧠 <b>RAM:</b>\n"
	if vm, err := mem.VirtualMemory(); err == nil {
		text += fmt.Sprintf("• Total: <code>%.2f GB</code>\n", float64(vm.Total)/1e9)
		text += fmt.Sprintf("• Used: <code>%.2f GB</code>\n", float64(vm.Used)/1e9)
		text += fmt.Sprintf("• Free: <code>%.2f GB</code>\n", float64(vm.Available)/1e9)
		text += fmt.Sprintf("• Usage: <code>%.2f%%</code>\n", vm.UsedPercent)
	}

	text += "\n💾 <b>Disk:</b>\n"
	if d, err := disk.Usage("/"); err == nil {
		text += fmt.Sprintf("• Total: <code>%.2f GB</code>\n", float64(d.Total)/1e9)
		text += fmt.Sprintf("• Used: <code>%.2f GB</code>\n", float64(d.Used)/1e9)
		text += fmt.Sprintf("• Free: <code>%.2f GB</code>\n", float64(d.Free)/1e9)
		text += fmt.Sprintf("• Usage: <code>%.2f%%</code>\n", d.UsedPercent)
	}

	text += "\n🔧 <b>System Info:</b>\n"
	text += fmt.Sprintf("• Go Version: <code>%s</code>\n", runtime.Version())
	text += fmt.Sprintf("• OS: <code>%s</code>\n", runtime.GOOS)
	text += fmt.Sprintf("• Arch: <code>%s</code>\n", runtime.GOARCH)
	text += fmt.Sprintf("• CPUs: <code>%d</code>\n", runtime.NumCPU())
	text += fmt.Sprintf("• Goroutines: <code>%d</code>\n", runtime.NumGoroutine())

	if percent, err := cpu.Percent(0, false); err == nil && len(percent) > 0 {
		text += fmt.Sprintf("• CPU Usage: <code>%.2f%%</code>\n", percent[0])
	}

	if avg, err := load.Avg(); err == nil {
		text += fmt.Sprintf("• Load Avg (1m,5m,15m): <code>%.2f, %.2f, %.2f</code>\n", avg.Load1, avg.Load5, avg.Load15)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	text += fmt.Sprintf("• Go Alloc Mem: <code>%.2f MB</code>\n", float64(m.Alloc)/1024/1024)

	msg.Respond(text)

	return telegram.EndGroup
}
