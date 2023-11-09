package alarm

type IAlarm interface {
	SendAlarm(content string, limit ...interface{})
	SetWebhook(w string)
}

func GetAlarmInstance() IAlarm {
	return newWeChatAlarm()
}
