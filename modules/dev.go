package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func ShellHandle(m *telegram.NewMessage) error {
	cmd := m.Args()
	var cmd_args []string
	if cmd == "" {
		m.Reply("No command provided")
		return nil
	}

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		cmd_args_b := strings.Split(m.Args(), " ")
		cmd_args = []string{"/C"}
		cmd_args = append(cmd_args, cmd_args_b...)
	} else {
		cmd = strings.Split(cmd, " ")[0]
		cmd_args = strings.Split(m.Args(), " ")
		cmd_args = append(cmd_args[:0], cmd_args[1:]...)
	}
	cmx := exec.Command(cmd, cmd_args...)
	var out bytes.Buffer
	cmx.Stdout = &out
	var errx bytes.Buffer
	cmx.Stderr = &errx
	err := cmx.Run()

	if errx.String() == "" && out.String() == "" {
		if err != nil {
			m.Reply("<code>Error:</code> <b>" + err.Error() + "</b>")
			return nil
		}
		m.Reply("<code>No Output</code>")
		return nil
	}

	if out.String() != "" {
		m.Reply(`<pre lang="bash">` + strings.TrimSpace(out.String()) + `</pre>`)
	} else {
		m.Reply(`<pre lang="bash">` + strings.TrimSpace(errx.String()) + `</pre>`)
	}
	return nil
}

// --------- Eval function ------------

const boiler_code_for_eval = `
package main

import "fmt"
import "github.com/amarnathcjd/gogram/telegram"
import "encoding/json"

%s

var msg_id int32 = %d

var client *telegram.Client
var message *telegram.NewMessage
var m *telegram.NewMessage
var r *telegram.NewMessage
` + "var msg = `%s`\nvar snd = `%s`\nvar cht = `%s`\nvar chn = `%s`\nvar cch = `%s`" + `


func evalCode() {
        %s
}

func main() {
        var msg_o *telegram.MessageObj
        var snd_o *telegram.UserObj
        var cht_o *telegram.ChatObj
        var chn_o *telegram.Channel
        json.Unmarshal([]byte(msg), &msg_o)
        json.Unmarshal([]byte(snd), &snd_o)
        json.Unmarshal([]byte(cht), &cht_o)
        json.Unmarshal([]byte(chn), &chn_o)
        client, _ = telegram.NewClient(telegram.ClientConfig{
                StringSession: "%s",
        })

        client.Cache.ImportJSON([]byte(cch))

        client.Conn()

        x := []telegram.User{}
        y := []telegram.Chat{}
        x = append(x, snd_o)
        if chn_o != nil {
                y = append(y, chn_o)
        }
        if cht_o != nil {
                y = append(y, cht_o)
        }
        client.Cache.UpdatePeersToCache(x, y)
        idx := 0
        if cht_o != nil {
                idx = int(cht_o.ID)
        }
        if chn_o != nil {
                idx = int(chn_o.ID)
        }
        if snd_o != nil && idx == 0 {
                idx = int(snd_o.ID)
        }

        messageX, err := client.GetMessages(idx, &telegram.SearchOption{
                IDs: int(msg_id),
        })

        if err != nil {
                fmt.Println(err)
        }

        message = &messageX[0]
        m = message
        r, _ = message.GetReplyMessage()

        fmt.Println("output-start")
        evalCode()
}

func packMessage(c *telegram.Client, message telegram.Message, sender *telegram.UserObj, channel *telegram.Channel, chat *telegram.ChatObj) *telegram.NewMessage {
        var (
                m = &telegram.NewMessage{}
        )
        switch message := message.(type) {
        case *telegram.MessageObj:
                m.ID = message.ID
                m.OriginalUpdate = message
                m.Message = message
                m.Client = c
        default:
                return nil
        }
        m.Sender = sender
        m.Chat = chat
        m.Channel = channel
        if m.Channel != nil && (m.Sender.ID == m.Channel.ID) {
                m.SenderChat = channel
        } else {
                m.SenderChat = &telegram.Channel{}
        }
        m.Peer, _ = c.GetSendablePeer(message.(*telegram.MessageObj).PeerID)

        /*if m.IsMedia() {
                FileID := telegram.PackBotFileID(m.Media())
                m.File = &telegram.CustomFile{
                        FileID: FileID,
                        Name:   getFileName(m.Media()),
                        Size:   getFileSize(m.Media()),
                        Ext:    getFileExt(m.Media()),
                }
        }*/
        return m
}
`

func resolveImports(code string) (string, []string) {
	var imports []string
	importsRegex := regexp.MustCompile(`import\s*\(([\s\S]*?)\)|import\s*\"([\s\S]*?)\"`)
	importsMatches := importsRegex.FindAllStringSubmatch(code, -1)
	for _, v := range importsMatches {
		if v[1] != "" {
			imports = append(imports, v[1])
		} else {
			imports = append(imports, v[2])
		}
	}
	code = importsRegex.ReplaceAllString(code, "")
	return code, imports
}

func EvalHandle(m *telegram.NewMessage) error {
	code := strings.TrimSpace(strings.Join(strings.SplitN(m.RawText(), " ", 2)[1:], " "))
	code, imports := resolveImports(code)

	if code == "" {
		return nil
	}

	defer os.Remove("tmp/eval.go")
	defer os.Remove("tmp/eval_out.txt")
	defer os.Remove("tmp")

	resp, isfile := perfomEval(code, m, imports)
	if isfile {
		if _, err := m.ReplyMedia(resp, telegram.MediaOptions{Caption: "Output"}); err != nil {
			m.Reply("Error: " + err.Error())
		}
		return nil
	}
	resp = strings.TrimSpace(resp)

	if resp != "" {
		if _, err := m.Reply(resp); err != nil {
			m.Reply(err)
		}
	}
	return nil
}

func perfomEval(code string, m *telegram.NewMessage, imports []string) (string, bool) {
	msg_b, _ := json.Marshal(m.Message)
	snd_b, _ := json.Marshal(m.Sender)
	cnt_b, _ := json.Marshal(m.Chat)
	chn_b, _ := json.Marshal(m.Channel)
	cache_b, _ := m.Client.Cache.ExportJSON()
	var importStatement string = ""
	if len(imports) > 0 {
		importStatement = "import (\n"
		for _, v := range imports {
			importStatement += `"` + v + `"` + "\n"
		}
		importStatement += ")\n"
	}

	code_file := fmt.Sprintf(boiler_code_for_eval, importStatement, m.ID, msg_b, snd_b, cnt_b, chn_b, cache_b, code, m.Client.ExportSession())
	tmp_dir := "tmp"
	_, err := os.ReadDir(tmp_dir)
	if err != nil {
		err = os.Mkdir(tmp_dir, 0o755)
		if err != nil {
			fmt.Println(err)
		}
	}

	// defer os.Remove(tmp_dir)

	os.WriteFile(tmp_dir+"/eval.go", []byte(code_file), 0o644)
	cmd := exec.Command("go", "run", "tmp/eval.go")
	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	err = cmd.Run()
	if stdOut.String() == "" && stdErr.String() == "" {
		if err != nil {
			return fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", err.Error()), false
		}
		return "<b>#EVALOut:</b> <code>No Output</code>", false
	}

	if stdOut.String() != "" {
		if len(stdOut.String()) > 4095 {
			os.WriteFile("tmp/eval_out.txt", stdOut.Bytes(), 0o644)
			return "tmp/eval_out.txt", true
		}

		strDou := strings.Split(stdOut.String(), "output-start")

		return fmt.Sprintf("<b>#EVALOut:</b> <code>%s</code>", strings.TrimSpace(strDou[1])), false
	}

	if stdErr.String() != "" {
		regexErr := regexp.MustCompile(`eval.go:\d+:\d+:`)
		errMsg := regexErr.Split(stdErr.String(), -1)
		if len(errMsg) > 1 {
			if len(errMsg[1]) > 4095 {
				os.WriteFile("tmp/eval_out.txt", []byte(errMsg[1]), 0o644)
				return "tmp/eval_out.txt", true
			}
			return fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", strings.TrimSpace(errMsg[1])), false
		}
		return fmt.Sprintf("<b>#EVALERR:</b> <code>%s</code>", stdErr.String()), false
	}

	return "<b>#EVALOut:</b> <code>No Output</code>", false
}

func LsHandler(m *telegram.NewMessage) error {
	dir := m.Args()
	if dir == "" {
		dir = "."
	}
	cmd := exec.Command("ls", dir)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	fileTypeEmoji := map[string]string{
		"file":   "📄",
		"dir":    "📁",
		"video":  "🎥",
		"audio":  "🎵",
		"image":  "🖼️",
		"go":     "📜",
		"python": "🐍",
		"txt":    "📝",
	}

	if err != nil {
		m.Reply("<code>Error:</code> <b>" + err.Error() + "</b>")
		return nil
	}

	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	var sizeTotal int64

	var resp string
	for _, file := range files {
		fileType := "file"
		if strings.Contains(file, ".") {
			fp := strings.Split(file, ".")
			fileType = fp[len(fp)-1]
		}
		switch fileType {
		case "mp4", "mkv", "webm", "avi", "flv", "mov", "wmv", "3gp":
			fileType = "video"
		case "mp3", "wav", "flac", "ogg", "m4a", "wma":
			fileType = "audio"
		case "jpg", "jpeg", "png", "gif", "webp", "bmp", "tiff":
			fileType = "image"
		case "go":
			fileType = "go"
		case "py":
			fileType = "python"
		case "txt":
			fileType = "txt"
		default:
			fileType = "file"
		}
		size := calcFileOrDirSize(filepath.Join(dir, file))
		sizeTotal += size
		resp += fileTypeEmoji[fileType] + " " + file + " " + "(" + sizeToHuman(size) + ")" + "\n"
	}

	resp += "\nTotal: " + sizeToHuman(sizeTotal)

	m.Reply("<pre lang='bash'>" + resp + "</pre>")
	return nil
}

func sizeToHuman(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

func calcFileOrDirSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}

	if !fi.IsDir() {
		return fi.Size()
	}

	var size int64
	walker := func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fi, err := info.Info()
			if err != nil {
				return err
			}
			size += fi.Size()
		}
		return nil
	}

	err = filepath.WalkDir(path, walker)
	if err != nil {
		return 0
	}

	return size
}
