package vklp

// 0,$message_id,0 -- delete a message with the local_id indicated
type DeleteMessage struct {
	MessageID uint64
}

// 1,$message_id,$flags -- replace message flags (FLAGS:=$flags)
type ReplaceMessageFlags struct {
	MessageID uint64
	Flags     uint64
}

// 2,$message_id,$mask[,$user_id] -- install message flags (FLAGS|=$mask)
type InstallMessageFlags struct {
	MessageID uint64
	Mask      uint64
	UserID    uint64
}

// 3,$message_id,$mask[,$user_id] -- reset message flags (FLAGS&=~$mask)
type ResetMessageFlags struct {
	MessageID uint64
	Mask      uint64
	UserID    uint64
}

// 4,$message_id,$flags,$from_id,$timestamp,$subject,$text,$attachments -- add a new message
type AddNewMessage struct {
	MessageID   uint64
	Flags       uint64
	FromID      uint64
	Timestamp   uint64
	Subject     string
	Text        string
	Attachments map[string]string
}

// 6,$peer_id,$local_id -- read all incoming messages with $peer_id until $local_id
type ReadAllIncomingMessages struct {
	PeerID  uint64
	LocalID uint64
}

// 7,$peer_id,$local_id -- read all outgoing messages with $peer_id until $local_id
type ReadAllOutgoingMessages struct {
	PeerID  uint64
	LocalID uint64
}

// 8,-$user_id,$extra -- a friend of $user_id is online, $extra is not 0 if flag 64 was transmitted in mode. $extra mod 256 is a platform id (see the list below)
type FriendOnline struct {
	UserID uint64
	Extra  uint64
}

// 9,-$user_id,$flags -- a friend of $user_id is offline ($flags equals 0 if the user has left the site (for example, clicked on "Log Out"), and 1 if offline upon timeout (for example, the status is set to "away"))
type FriendOffline struct {
	UserID uint64
	Flags  uint64
}

// 51,$chat_id,$self -- one of $chat_id's parameters (title, participants) was changed. $self shows if changes were made by user themself
type ChatParamsChanged struct {
	ChatID uint64
	Self   uint8
}

// 61,$user_id,$flags -- $user_id started typing text in a dialog. The event is sent once in ~5 sec while constantly typing. $flags = 1
type StartedTypingInDialog struct {
	UserID uint64
	Flags  uint64
}

// 62,$user_id,$chat_id — $user_id started typing in $chat_id.
type StartedTypingInChat struct {
	UserID uint64
	ChatID uint64
}

// 70,$user_id,$call_id — $user_id made a call with $call_id identifier.
type Call struct {
	UserID uint64
	CallID uint64
}

// 80,$count,0 — new unread messages counter in the left menu equals $count.
type UnreadMessages struct {
	Count uint64
}

// 114,{ $peerId, $sound, $disabled_until } — notification settings changed, where peerId is a chat's/user's $peer_id, sound — 1 || 0, sound notifications on/off, disabled_until — notifications disabled for a certain period (-1: forever; 0: notifications enabled; other: timestamp for time to switch back on).
type NotificationSettingsChanged struct {
	PeerID        uint64
	Sound         uint8
	DisabledUntil uint64
}
