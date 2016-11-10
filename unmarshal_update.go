package vklp

import (
	"encoding/json"
	"fmt"
)

func unmarshalUpdate(b []byte) (interface{}, error) {
	r := readerPool.For(b)
	d := json.NewDecoder(r)

	if t, err := d.Token(); err != nil {
		return nil, err
	} else if delim, ok := t.(json.Delim); !ok {
		return nil, fmt.Errorf("vklp: expected first json token to be delimiter, got %#v", t)
	} else if delim != json.Delim('[') {
		return nil, fmt.Errorf("vklp: expected first json delimiter to be array opening bracket ('['), got '%s'", delim.String())
	}

	var tp uint8
	if err := d.Decode(&tp); err != nil {
		return nil, err
	}

	switch tp {

	case 0:
		// 0,$message_id,0 -- delete a message with the local_id indicated
		u := new(DeleteMessage)
		if err := unmarshal(d, &u.MessageID); err != nil {
			return nil, err
		}
		return u, nil

	case 1:
		// 1,$message_id,$flags -- replace message flags (FLAGS:=$flags)
		u := new(ReplaceMessageFlags)
		if err := unmarshal(d, &u.MessageID, &u.Flags); err != nil {
			return nil, err
		}
		return u, nil

	case 2:
		// 2,$message_id,$mask[,$user_id] -- install message flags (FLAGS|=$mask)
		u := new(InstallMessageFlags)
		if err := unmarshal(d, &u.MessageID, &u.Mask, &u.UserID); err != nil {
			return nil, err
		}
		return u, nil

	case 3:
		// 3,$message_id,$mask[,$user_id] -- reset message flags (FLAGS&=~$mask)
		u := new(ResetMessageFlags)
		if err := unmarshal(d, &u.MessageID, &u.Mask, &u.UserID); err != nil {
			return nil, err
		}
		return u, nil

	case 4:
		// 4,$message_id,$flags,$from_id,$timestamp,$subject,$text,$attachments -- add a new message
		u := new(AddNewMessage)
		if err := unmarshal(d, &u.MessageID, &u.Flags, &u.FromID, &u.Timestamp, &u.Subject, &u.Text, &u.Attachments); err != nil {
			return nil, err
		}
		return u, nil

	case 6:
		// 6,$peer_id,$local_id -- read all incoming messages with $peer_id until $local_id
		u := new(ReadAllIncomingMessages)
		if err := unmarshal(d, &u.PeerID, &u.LocalID); err != nil {
			return nil, err
		}
		return u, nil

	case 7:
		// 7,$peer_id,$local_id -- read all outgoing messages with $peer_id until $local_id
		u := new(ReadAllOutgoingMessages)
		if err := unmarshal(d, &u.PeerID, &u.LocalID); err != nil {
			return nil, err
		}
		return u, nil

	case 8:
		// 8,-$user_id,$extra -- a friend of $user_id is online, $extra is not 0 if flag 64 was transmitted in mode. $extra mod 256 is a platform id (see the list below)
		u := new(FriendOnline)
		var userID int64
		if err := unmarshal(d, &userID, &u.Extra); err != nil {
			return nil, err
		}
		u.UserID = uint64(-userID)
		return u, nil

	case 9:
		// 9,-$user_id,$flags -- a friend of $user_id is offline ($flags equals 0 if the user has left the site (for example, clicked on "Log Out"), and 1 if offline upon timeout (for example, the status is set to "away"))
		u := new(FriendOffline)
		var userID int64
		if err := unmarshal(d, &userID, &u.Flags); err != nil {
			return nil, err
		}
		u.UserID = uint64(-userID)
		return u, nil

	case 51:
		// 51,$chat_id,$self -- one of $chat_id's parameters (title, participants) was changed. $self shows if changes were made by user themself
		u := new(ChatParamsChanged)
		if err := unmarshal(d, &u.ChatID, &u.Self); err != nil {
			return nil, err
		}
		return u, nil

	case 61:
		// 61,$user_id,$flags -- $user_id started typing text in a dialog. The event is sent once in ~5 sec while constantly typing. $flags = 1
		u := new(StartedTypingInDialog)
		if err := unmarshal(d, &u.UserID, &u.Flags); err != nil {
			return nil, err
		}
		return u, nil

	case 62:
		// 62,$user_id,$chat_id — $user_id started typing in $chat_id.
		u := new(StartedTypingInChat)
		if err := unmarshal(d, &u.UserID, &u.ChatID); err != nil {
			return nil, err
		}
		return u, nil

	case 70:
		// 70,$user_id,$call_id — $user_id made a call with $call_id identifier.
		u := new(Call)
		if err := unmarshal(d, &u.UserID, &u.CallID); err != nil {
			return nil, err
		}
		return u, nil

	case 80:
		// 80,$count,0 — new unread messages counter in the left menu equals $count.
		u := new(UnreadMessages)
		if err := unmarshal(d, &u.Count); err != nil {
			return nil, err
		}
		return u, nil

	case 114:
		// 114,{ $peerId, $sound, $disabled_until } — notification settings changed, where peerId is a chat's/user's $peer_id, sound — 1 || 0, sound notifications on/off, disabled_until — notifications disabled for a certain period (-1: forever; 0: notifications enabled; other: timestamp for time to switch back on).
		u := new(NotificationSettingsChanged)
		if err := unmarshal(d, &u.PeerID, &u.Sound, &u.DisabledUntil); err != nil {
			return nil, err
		}
		return u, nil

	default:
		return nil, fmt.Errorf("vklp: unsupported update type %d", tp)

	}
}

func unmarshal(d *json.Decoder, vs ...interface{}) error {
	for i := 0; i < len(vs); i++ {
		if !d.More() {
			return nil
		}
		if err := d.Decode(vs[i]); err != nil {
			return err
		}
	}
	return nil
}
