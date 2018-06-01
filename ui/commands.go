// gomuks - A terminal Matrix client written in Go.
// Copyright (C) 2018 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ui

import (
	"maunium.net/go/gomuks/debug"
	"strings"
	"encoding/json"
)

func cmdMe(cmd *Command) {
	text := strings.Join(cmd.Args, " ")
	tempMessage := cmd.Room.NewTempMessage("m.emote", text)
	go cmd.MainView.sendTempMessage(cmd.Room, tempMessage, text)
	cmd.UI.Render()
}

func cmdQuit(cmd *Command) {
	cmd.Gomuks.Stop()
}

func cmdClearCache(cmd *Command) {
	cmd.Config.Clear()
	cmd.Gomuks.Stop()
}

func cmdUnknownCommand(cmd *Command) {
	cmd.Reply("Unknown command \"%s\". Try \"/help\" for help.", cmd.Command)
}

func cmdHelp(cmd *Command) {
	cmd.Reply("Known command. Don't try \"/help\" for help.")
}

func cmdLeave(cmd *Command) {
	err := cmd.Matrix.LeaveRoom(cmd.Room.MxRoom().ID)
	debug.Print("Leave room error:", err)
	if err == nil {
		cmd.MainView.RemoveRoom(cmd.Room.MxRoom())
	}
}

func cmdJoin(cmd *Command) {
	if len(cmd.Args) == 0 {
		cmd.Reply("Usage: /join <room>")
		return
	}
	identifer := cmd.Args[0]
	server := ""
	if len(cmd.Args) > 1 {
		server = cmd.Args[1]
	}
	room, err := cmd.Matrix.JoinRoom(identifer, server)
	debug.Print("Join room error:", err)
	if err == nil {
		cmd.MainView.AddRoom(room)
	}
}

func cmdSendEvent(cmd *Command) {
	debug.Print(cmd.Command, cmd.Args, len(cmd.Args))
	if len(cmd.Args) < 3 {
		cmd.Reply("Usage: /send <room id> <event type> <content>")
		return
	}
	roomID := cmd.Args[0]
	eventType := cmd.Args[1]
	rawContent := strings.Join(cmd.Args[2:], "")
	debug.Print(roomID, eventType, rawContent)

	var content interface{}
	err := json.Unmarshal([]byte(rawContent), &content)
	debug.Print(err)
	if err != nil {
		cmd.Reply("Failed to parse content: %v", err)
		return
	}
	debug.Print("Sending event to", roomID, eventType, content)

	resp, err := cmd.Matrix.Client().SendMessageEvent(roomID, eventType, content)
	debug.Print(resp, err)
	if err != nil {
		cmd.Reply("Error from server: %v", err)
	} else {
		cmd.Reply("Event sent, ID: %s", resp.EventID)
	}
}

func cmdSetState(cmd *Command) {
	if len(cmd.Args) < 4 {
		cmd.Reply("Usage: /setstate <room id> <event type> <state key/`-`> <content>")
		return
	}

	roomID := cmd.Args[0]
	eventType := cmd.Args[1]
	stateKey := cmd.Args[2]
	if stateKey == "-" {
		stateKey = ""
	}
	rawContent := strings.Join(cmd.Args[3:], "")

	var content interface{}
	err := json.Unmarshal([]byte(rawContent), &content)
	if err != nil {
		cmd.Reply("Failed to parse content: %v", err)
		return
	}
	debug.Print("Sending state event to", roomID, eventType, stateKey, content)
	resp, err := cmd.Matrix.Client().SendStateEvent(roomID, eventType, stateKey, content)
	if err != nil {
		cmd.Reply("Error from server: %v", err)
	} else {
		cmd.Reply("State event sent, ID: %s", resp.EventID)
	}
}

func cmdUIToggle(cmd *Command) {
	if len(cmd.Args) == 0 {
		cmd.Reply("Usage: /uitoggle <rooms/users/baremessages/images>")
		return
	}
	switch cmd.Args[0] {
	case "rooms":
		cmd.Config.Preferences.HideRoomList = !cmd.Config.Preferences.HideRoomList
	case "users":
		cmd.Config.Preferences.HideUserList = !cmd.Config.Preferences.HideUserList
	case "baremessages":
		cmd.Config.Preferences.BareMessageView = !cmd.Config.Preferences.BareMessageView
	case "images":
		cmd.Config.Preferences.DisableImages = !cmd.Config.Preferences.DisableImages
	default:
		cmd.Reply("Usage: /uitoggle <rooms/users/baremessages/images>")
		return
	}
	cmd.UI.Render()
	cmd.UI.Render()
	go cmd.Matrix.SendPreferencesToMatrix()
}

func cmdLogout(cmd *Command) {
	cmd.Matrix.Logout()
}
