package modules

func BroadcastFunc(m *telegram.NewMessage) error {

userChan, chatChan, err := m.Client.Broadcast()
if err != nil {
    log.Println(err)
    m Reply(err.Error())
return telegram.EndGroup
}


userCount := 0
for user := range userChan {
    userCount++
}

chatCount := 0
for chat := range chatChan {
    chatCount++
}
m.Reply(fmt.Sprinf("Total Chats: %d\nTotal Users: %d", chatCount, chatCount))

}