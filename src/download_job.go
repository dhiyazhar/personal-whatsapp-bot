package main

import (
	"go.mau.fi/whatsmeow/types"
)

type DownloadJob struct {
	UserInfo     types.MessageInfo
	VideoURL     string
	InitialMsgID types.MessageID
}
