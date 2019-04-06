/*
This package is the embeded version of 'github.com/Sioro-Neoku/go-peerflix/'.
We did some modifications on it in order to let it fit into 'torrodle'
*/
package player

import (
	"github.com/sirupsen/logrus"
	"os/exec"
	"runtime"
	"strings"
)

// Players holds structs of all supported players.
var Players = []Player{
	{
		Name:            "MPV",
		DarwinCommand:   []string{"mpv"},
		LinuxCommand:    []string{"mpv"},
		WindowsCommand:  []string{"mpv"},
		SubtitleCommand: "--sub-file=",
	},
	{
		Name:            "VLC",
		DarwinCommand:   []string{"open", "-a", "vlc"},
		LinuxCommand:    []string{"vlc"},
		WindowsCommand:  []string{"%PROGRAMFILES%\\VideoLAN\\VLC\\vlc.exe"},
		SubtitleCommand: "--sub-file=",
	},
}

// Player manages the execiution of a media player.
type Player struct {
	Name            string
	DarwinCommand   []string
	LinuxCommand    []string
	WindowsCommand  []string
	SubtitleCommand string
	started         bool
}

// Start launches the Player with the given command and arguments in subprocess.
func (player *Player) Start(url string, subtitlePath string) {
	if player.started == true {
		// prevent multiple calls
		return
	}
	var command []string
	switch runtime.GOOS {
	case "darwin":
		command = player.DarwinCommand
	case "linux":
		command = player.LinuxCommand
	case "windows":
		command = player.WindowsCommand
	}
	command = append(command, url)
	if subtitlePath != "" {
		command = append(command, player.SubtitleCommand + subtitlePath)
	}
	logrus.Debugf("command: %v\n", command)
	cmd := exec.Command(command[0], command[1:]...)
	player.started = true
	go cmd.Start()
}

// GetPlayer returns the Player struct of the given player name.
func GetPlayer(name string) *Player {
	for _, player := range Players {
		if strings.ToLower(player.Name) == strings.ToLower(name) {
			return &player
		}
	}
	return nil
}
