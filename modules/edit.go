package modules

import (
	"github.com/amarnathcjd/gogram/telegram"

	"main/config/helpers"
)

func deleteEditedMessage(m *telegram.NewMessage) error {
	if !IsSupergroup(m) || m.Message.EditHide {
		return nil
	}
	if ShouldIgnoreGroupAnonymous(m) {
		return nil
	}
	if isadmin, err := helpers.IsChatAdmin(m.Client, m.ChannelID(), m.Sender.ID); err != nil {
		return err
	} else if isadmin {
		return nil
	}

	if _, err := m.Delete(); err != nil && handleNeedPerm(err, m) {
		return telegram.EndGroup
	}

	reason := "<b>🚫 Editing messages is prohibited in this chat.</b> Please refrain from modifying your messages to maintain the integrity of the conversation."
	p := "<b>📷 Photo edits are blocked.</b> Images must stay unchanged to preserve context."

	v := "<b>🎥 Video edits aren't allowed.</b> Videos must remain as originally shared."

	d := "<b>📄 Document edits are restricted.</b> Keep documents unchanged for reliability."

	a := "<b>🎵 Audio edits aren't permitted.</b> Audio files must remain unaltered."

	switch {
	case m.Text() != "" && !m.IsMedia():
		reason = "<b>🚫 Editing text is not allowed.</b> Please avoid changing messages once sent to keep conversations clear."

	case m.Text() != "" && m.IsMedia():
		reason = "<b>✍️ Caption edits are restricted.</b> Changing them affects clarity and is not permitted."
		if m.Photo() != nil {
			reason = p
		} else if m.Video() != nil {
			reason = v
		} else if m.Document() != nil {
			reason = d
		} else if m.Audio() != nil {
			reason = a
		}

	case m.Photo() != nil:
		reason = p

	case m.Video() != nil:
		reason = v

	case m.Document() != nil:
		reason = d

	case m.Audio() != nil:
		reason = a

	case m.Voice() != nil:
		reason = "<b>🎙️ Voice edits are restricted.</b> Voice messages should remain original."

	case m.Animation() != nil:
		reason = "<b>🎞️ GIF edits are blocked.</b> Keep animations unchanged for context."

	}

	_, err := m.Respond(reason)

	return orCont(err)
}
