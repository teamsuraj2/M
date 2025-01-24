package helpers

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

func GetFullChannel(c *telegram.Client, chatId any) (*telegram.ChannelFull, error) {
	resolvedPeer, err := c.ResolvePeer(chatId)
	if err != nil {
		return nil, err
	}

	inPeer, ok := resolvedPeer.(*telegram.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf("chatId is not a channel")
	}

	fullChatRaw, err := c.ChannelsGetFullChannel(&telegram.InputChannelObj{
		ChannelID:  inPeer.ChannelID,
		AccessHash: inPeer.AccessHash,
	})
	if err != nil {
		return nil, err
	}

	if fullChatRaw == nil {
		return nil, fmt.Errorf("fullChatRaw is nil")
	}

	fullChat, ok := fullChatRaw.FullChat.(*telegram.ChannelFull)
	if !ok {
		return nil, fmt.Errorf("fullChatRaw.FullChat is not a ChannelFull")
	}

	return fullChat, nil
}
