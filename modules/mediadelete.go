package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config/helpers"
	"main/database"
)

func init() {
	AddHelp(
		"🎬 MediaDelete",
		"mediadelete_help",
		"🎬 <b>Media Auto-Delete</b> automatically removes media messages after a set delay.\n\n"+
			"<b>Commands:</b>\n"+
			"• <code>/mediadelete on</code> — Enable media auto-delete ✅\n"+
			"• <code>/mediadelete off</code> — Disable media auto-delete ❌\n"+
			"• <code>/setmediadelay &lt;time&gt;</code> — Set delay (1m to 24h)\n"+
			"• <code>/mediaexcept &lt;user/admin/off&gt;</code> — Configure deletion mode (owner only)\n\n"+
			"<b>Time Format:</b>\n"+
			"• <code>5m</code> = 5 minutes\n"+
			"• <code>1h</code> = 1 hour\n"+
			"• <code>12h</code> = 12 hours\n\n"+
			"<b>Media Except Modes:</b>\n"+
			"• <code>user</code> or <code>on</code> — Delete media ONLY from regular users (admins safe)\n"+
			"• <code>admin</code> — Delete media from BOTH users AND admins\n"+
			"• <code>off</code> — Don't delete any media\n\n"+
			"<b>ℹ️ Note:</b> Applies to photos, videos, stickers, GIFs, and animations.\n"+
			"👮 Only group admins can configure this setting.\n"+
			"👑 Only group owner can set exception mode.",
	)
}

func MediaDeleteCmd(m *telegram.NewMessage) error {
	args := strings.Fields(m.Text())
	if isgroup := IsValidSupergroup(m); !isgroup {
		return telegram.EndGroup
	}
	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	} else if err != nil {
		return L(m, "Modules -> mediadelete -> m.Delete()", err)
	}
	if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChatID(), m.SenderID()); err != nil {
		return L(m, "Modules -> mediadelete -> helpers.IsChatAdmin()", err)
	} else if !isadmin {
		m.Respond("Access denied: Only group admins can use this command.")
		return telegram.EndGroup
	}

	if len(args) < 2 {
		settings, err := database.GetMediaDeleteSettings(m.ChatID())
		if err != nil {
			m.Respond("⚠️ Usage:\n<code>/mediadelete on</code> — Enable\n<code>/mediadelete off</code> — Disable")
		} else {
			status := map[bool]string{true: "Enabled", false: "Disabled"}[settings.Enabled]
			mode := "off"
			if settings.DeleteFromUsers {
				mode = "user (delete only from users)"
			}
			if settings.DeleteFromAdmins {
				mode = "admin (delete from both users and admins)"
			}
			m.Respond(fmt.Sprintf("📊 <b>MediaDelete Status:</b> %s\n⏱ <b>Delay:</b> %s\n🎯 <b>Mode:</b> %s", status, formatDuration(settings.Delay), mode))
		}
		return telegram.EndGroup
	}

	arg := strings.ToLower(args[1])
	enable := arg == "on"
	if arg != "on" && arg != "off" {
		m.Respond("❌ Invalid option.\nUse <code>/mediadelete on</code> or <code>/mediadelete off</code>")
		return telegram.EndGroup
	}

	if err := database.SetMediaDeleteEnabled(m.ChatID(), enable); err != nil {
		m.Respond("❌ Failed to update setting.")
		return L(m, "Modules -> mediadelete -> SetMediaDeleteEnabled", err)
	}

	status := "🎬 Media auto-delete enabled ✅"
	if !enable {
		status = "🚫 Media auto-delete disabled ❌"
	}
	m.Respond(status)
	return telegram.EndGroup
}

func SetMediaDelayCmd(m *telegram.NewMessage) error {
	if isgroup := IsValidSupergroup(m); !isgroup {
		return telegram.EndGroup
	}
	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	} else if err != nil {
		return L(m, "Modules -> mediadelete -> m.Delete()", err)
	}
	if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChatID(), m.SenderID()); err != nil {
		return L(m, "Modules -> mediadelete -> helpers.IsChatAdmin()", err)
	} else if !isadmin {
		m.Respond("Access denied: Only group admins can use this command.")
		return telegram.EndGroup
	}

	if m.Args() == "" {
		m.Respond("⚠️ Usage: <code>/setmediadelay 5m</code> or <code>/setmediadelay 2h</code>\n\nValid range: 1m to 24h")
		return telegram.EndGroup
	}

	duration, err := parseDuration(m.Args())
	if err != nil {
		m.Respond("❌ Invalid time format. Use: 5m, 1h, 12h, etc.\nRange: 1 minute to 24 hours")
		return telegram.EndGroup
	}

	if duration < time.Minute || duration > 24*time.Hour {
		m.Respond("❌ Delay must be between 1 minute and 24 hours.")
		return telegram.EndGroup
	}

	if err := database.SetMediaDeleteDelay(m.ChatID(), duration); err != nil {
		m.Respond("❌ Failed to set delay.")
		return L(m, "Modules -> mediadelete -> SetMediaDeleteDelay", err)
	}

	m.Respond(fmt.Sprintf("✅ Media auto-delete delay set to %s", formatDuration(duration)))
	return telegram.EndGroup
}

func MediaExceptCmd(m *telegram.NewMessage) error {
	if isgroup := IsValidSupergroup(m); !isgroup {
		return telegram.EndGroup
	}
	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	} else if err != nil {
		return L(m, "Modules -> mediadelete -> m.Delete()", err)
	}

	// Check if user is owner
	owner, err := helpers.GetOwner(m.Client, m.ChatID())
	if err != nil {
		return L(m, "Modules -> mediadelete -> GetOwner", err)
	}
	if m.SenderID() != owner {
		m.Respond("🚫 Only the group owner can configure media deletion mode.")
		return telegram.EndGroup
	}

	if m.Args() == "" {
		settings, _ := database.GetMediaDeleteSettings(m.ChatID())
		current := "off"
		if settings.DeleteFromUsers && !settings.DeleteFromAdmins {
			current = "user (delete ONLY from users)"
		} else if settings.DeleteFromAdmins {
			current = "admin (delete from BOTH users AND admins)"
		}
		m.Respond(fmt.Sprintf("📋 <b>Current mode:</b> <code>%s</code>\n\n⚠️ <b>Usage:</b>\n<code>/mediaexcept user</code> — Delete media ONLY from users (admins safe)\n<code>/mediaexcept admin</code> — Delete from BOTH users AND admins\n<code>/mediaexcept off</code> — Don't delete any media", current))
		return telegram.EndGroup
	}

	mode := strings.ToLower(m.Args())

	var deleteFromUsers, deleteFromAdmins bool
	switch mode {
	case "user", "on":
		deleteFromUsers = true
		deleteFromAdmins = false
	case "admin":
		deleteFromUsers = true
		deleteFromAdmins = true
	case "off":
		deleteFromUsers = false
		deleteFromAdmins = false
	default:
		m.Respond("❌ Invalid mode. Use: <code>user</code>, <code>admin</code>, or <code>off</code>")
		return telegram.EndGroup
	}

	if err := database.SetMediaDeletionMode(m.ChatID(), deleteFromUsers, deleteFromAdmins); err != nil {
		m.Respond("❌ Failed to set mode.")
		return L(m, "Modules -> mediadelete -> SetMediaDeletionMode", err)
	}

	var msg string
	if deleteFromUsers && !deleteFromAdmins {
		msg = "✅ Media will be deleted from <b>USERS ONLY</b> (admins are safe)"
	} else if deleteFromAdmins {
		msg = "✅ Media will be deleted from <b>BOTH users AND admins</b>"
	} else {
		msg = "✅ Media deletion disabled - no media will be deleted"
	}

	m.Respond(msg)
	return telegram.EndGroup
}

func handleMediaDelete(m *telegram.NewMessage) error {
	if !IsSupergroup(m) || !m.IsMedia() {
		return nil
	}

	settings, err := database.GetMediaDeleteSettings(m.ChatID())
	if err != nil || !settings.Enabled {
		return nil
	}

	if ShouldIgnoreGroupAnonymous(m) {
		return nil
	}

	isadmin, err := helpers.IsChatAdmin(m.Client, m.ChatID(), m.Sender.ID)
	if err != nil {
		L(m, "Modules -> mediadelete -> helpers.IsChatAdmin()", err)
		return nil
	}

	// Check deletion mode
	shouldDelete := false

	if !settings.DeleteFromUsers && !settings.DeleteFromAdmins {
		// Mode: off - don't delete anything
		return nil
	} else if settings.DeleteFromUsers && !settings.DeleteFromAdmins {
		// Mode: user - delete only from regular users (admins safe)
		if !isadmin {
			shouldDelete = true
		}
	} else if settings.DeleteFromAdmins {
		// Mode: admin - delete from both users and admins
		shouldDelete = true
	}

	if !shouldDelete {
		return nil
	}

	// Schedule deletion using the global scheduler
	ScheduleMessageDeletion(m.Client, m.ChatID(), int32(m.ID), settings.Delay)

	return nil
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration")
	}

	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}

	switch unit {
	case 'm':
		return time.Duration(num) * time.Minute, nil
	case 'h':
		return time.Duration(num) * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid unit (use m or h)")
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
