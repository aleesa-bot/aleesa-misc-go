package main

// Конфиг
type myConfig struct {
	Server          string `json:"server,omitempty"`
	Port            int    `json:"port,omitempty"`
	Loglevel        string `json:"loglevel,omitempty"`
	Log             string `json:"log,omitempty"`
	Channel         string `json:"channel,omitempty"`
	ForwardChannels struct {
		Games   string `json:"games,omitempty"`
		Phrases string `json:"phrases,omitempty"`
		Webapp  string `json:"webapp,omitempty"`
		Craniac string `json:"craniac,omitempty"`
	} `json:"forward_channels"`
	Csign       string `json:"csign,omitempty"`
	ForwardsMax int64  `json:"forwards_max,omitempty"`
}

// Входящее сообщение из pubsub-канала redis-ки
type rMsg struct {
	From     string `json:"from,omitempty"`
	Chatid   string `json:"chatid,omitempty"`
	Userid   string `json:"userid,omitempty"`
	Threadid string `json:"threadid,omitempty"`
	Message  string `json:"message,omitempty"`
	Plugin   string `json:"plugin,omitempty"`
	Mode     string `json:"mode,omitempty"`
	Misc     struct {
		Answer      int64  `json:"answer,omitempty"`
		Botnick     string `json:"bot_nick,omitempty"`
		Csign       string `json:"csign,omitempty"`
		Fwdcnt      int64  `json:"fwd_cnt,omitempty"`
		GoodMorning int64  `json:"good_morning,omitempty"`
		Msgformat   int64  `json:"msg_format,omitempty"`
		Username    string `json:"username,omitempty"`
	} `json:"Misc"`
}

// Исходящее сообщение в pubsub-канал redis-ки
type sMsg struct {
	From     string `json:"from"`
	Chatid   string `json:"chatid"`
	Userid   string `json:"userid"`
	Threadid string `json:"threadid"`
	Message  string `json:"message"`
	Plugin   string `json:"plugin"`
	Mode     string `json:"mode"`
	Misc     struct {
		Answer      int64  `json:"answer"`
		Botnick     string `json:"bot_nick"`
		Csign       string `json:"csign"`
		Fwdcnt      int64  `json:"fwd_cnt"`
		GoodMorning int64  `json:"good_morning"`
		Msgformat   int64  `json:"msg_format"`
		Username    string `json:"username"`
	} `json:"misc"`
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */
