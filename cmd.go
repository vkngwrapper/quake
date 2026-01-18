package main

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

var CVarClNopext = CVar{
	Name:          "cl_nopext",
	DefaultString: "0",
}

var CVarClWarncmd = CVar{
	Name:          "cl_warncmd",
	DefaultString: "1",
}

type CmdSource int

const (
	CmdSourceClient CmdSource = iota
	CmdSourceCommand
	CmdSourceServer
)

const CmdMaxArgs int = 80
const AliasMaxNameLength int = 32

var Cmds *CmdExecutor = &CmdExecutor{}

type CmdCallbackFunc func()

type CmdFunction struct {
	Name     string
	Source   CmdSource
	Dynamic  bool
	Function CmdCallbackFunc
	Next     *CmdFunction
}

type CmdAlias struct {
	Name  string
	Value string
	Next  *CmdAlias
}

type CmdExecutor struct {
	functions *CmdFunction
	aliases   *CmdAlias
	waiting   bool
	buffer    []rune

	source    CmdSource
	argString string
	args      []string
}

func (e *CmdExecutor) Add(name string, function CmdCallbackFunc, source CmdSource) *CmdFunction {
	if CVars.FindVar(name) != nil {
		log.Printf("Cmd_AddCommand: %s already defined as a var\n", name)
		return nil
	}

	// Check if function already exists
	for cmd := e.functions; cmd != nil; cmd = cmd.Next {
		if cmd.Name == name && cmd.Source == source {
			if function != nil {
				log.Printf("Cmd_AddCommand: %s already defined\n", name)
			}
			return nil
		}
	}

	newCmd := &CmdFunction{
		Name:     name,
		Dynamic:  HostInitialized,
		Function: function,
		Source:   source,
	}

	if e.functions == nil || strings.Compare(newCmd.Name, e.functions.Name) < 0 {
		newCmd.Next = e.functions
		e.functions = newCmd
	} else {
		prev := e.functions
		cursor := e.functions.Next

		for cursor != nil && strings.Compare(newCmd.Name, cursor.Name) > 0 {
			prev = cursor
			cursor = cursor.Next
		}
		newCmd.Next = prev.Next
		prev.Next = newCmd
	}

	if newCmd.Dynamic {
		return newCmd
	}

	return nil
}

func (e *CmdExecutor) Remove(cmd *CmdFunction) {
	for link := &e.functions; *link != nil; link = &(*link).Next {
		if *link == cmd {
			*link = cmd.Next
			return
		}
	}

	log.Fatalf("Cmd_RemoveCommand: unable to remove command %s\n", cmd.Name)
}

func (e *CmdExecutor) Exists(cmdName string) bool {
	for cmd := e.functions; cmd != nil; cmd = cmd.Next {
		if cmd.Name == cmdName && cmd.Source == CmdSourceCommand {
			return true
		}
	}

	return false
}

func (e *CmdExecutor) CompleteCommandName(partial string) string {
	if partial == "" {
		return ""
	}

	for cmd := e.functions; cmd != nil; cmd = cmd.Next {
		if strings.HasPrefix(cmd.Name, partial) {
			return cmd.Name
		}
	}

	return ""
}

func (e *CmdExecutor) CmdWait() {
	e.waiting = true
}

func (e *CmdExecutor) Waited() {
	e.waiting = false
}

func (e *CmdExecutor) Init() {
	e.buffer = make([]rune, 0, 130000)

	e.Add("cmdlist", e.CmdList, CmdSourceCommand)
	e.Add("unalias", e.CmdUnalias, CmdSourceCommand)
	e.Add("unaliasall", e.CmdUnaliasAll, CmdSourceCommand)

	e.Add("stuffcmd", e.CmdStuffCmds, CmdSourceCommand)
	e.Add("exec", e.CmdExec, CmdSourceCommand)
	e.Add("echo", CmdEcho, CmdSourceCommand)
	e.Add("alias", e.CmdAlias, CmdSourceCommand)
	// TODO: Networking
	//e.Add("cmd", e.CmdForwardToServer, CmdSourceCommand)
	e.Add("wait", e.CmdWait, CmdSourceCommand)

	e.Add("apropos", e.CmdApropos, CmdSourceCommand)
	e.Add("find", e.CmdApropos, CmdSourceCommand)

	CVars.Register(&CVarClNopext)
	CVars.Register(&CVarClWarncmd)
}

func (e *CmdExecutor) AddText(text string) {
	for _, r := range text {
		e.buffer = append(e.buffer, r)
	}
}

func (e *CmdExecutor) InsertText(text string) {
	// Expand slice to cover new text size
	addedLen := len(text) + 1
	existingLen := len(e.buffer)
	if cap(e.buffer) > existingLen+addedLen {
		e.buffer = e.buffer[:existingLen+addedLen]
	} else {
		for i := 0; i < addedLen; i++ {
			e.buffer = append(e.buffer, '0')
		}
	}

	// Copy existing text to new position
	copy(e.buffer[addedLen:addedLen+existingLen], e.buffer[:existingLen])

	// Copy new text to beginning
	for index, r := range text {
		e.buffer[index] = r
	}
	e.buffer[addedLen-1] = '\n'
}

func (e *CmdExecutor) Execute() {
	for len(e.buffer) > 0 && !e.waiting {
		// Find a \n or ; line break
		quotes := 0
		comment := false
		var textIndex int
		for textIndex = 0; textIndex < len(e.buffer); textIndex++ {
			if e.buffer[textIndex] == '"' {
				quotes++
			}
			if e.buffer[textIndex] == '/' && len(e.buffer)-1 > textIndex && e.buffer[textIndex+1] == '/' {
				comment = true
			}
			if quotes%2 == 0 && !comment && e.buffer[textIndex] == ';' {
				// Encountered a semicolon not inside a quote and not inside a comment
				break
			}
			if e.buffer[textIndex] == '\n' {
				break
			}
		}

		line := string(e.buffer[:textIndex])

		// Delete line from buffer
		textIndex++
		remainingBuffer := len(e.buffer) - textIndex
		if remainingBuffer > 0 {
			copy(e.buffer[:remainingBuffer], e.buffer[textIndex:textIndex+remainingBuffer])
		} else if remainingBuffer < 0 {
			remainingBuffer = 0
		}
		e.buffer = e.buffer[:remainingBuffer]

		e.ExecuteString(line, CmdSourceCommand)
	}
}

func (e *CmdExecutor) Args() string {
	return e.argString
}

func (e *CmdExecutor) ArgCount() int {
	return len(e.args)
}

func (e *CmdExecutor) Arg(index int) string {
	if index < 0 || index >= len(e.args) {
		return ""
	}
	return e.args[index]
}

func (e *CmdExecutor) checkParam(param string) int {
	if param == "" {
		log.Fatalln("Cmd_CheckParm: empty input")
	}

	for i := 1; i < len(e.args); i++ {
		if strings.EqualFold(param, e.args[i]) {
			return i
		}
	}

	return 0
}

func (e *CmdExecutor) TokenizeBuffer(buffer []rune) {
	e.argString = ""
	e.args = e.args[:0]

	index := 0
	for {
		// Skip whitespace
		for index < len(buffer) && buffer[index] <= ' ' && buffer[index] != '\n' {
			index++
		}

		// Linebreak is end of command
		if buffer[index] == '\n' {
			// End of command
			index++
			break
		}

		if index >= len(buffer) {
			break
		}

		if len(e.args) == 1 {
			e.argString = string(buffer[index:])
		}

		token := ParseToken(buffer[index:])
		if token == "" {
			return
		}

		if len(e.args) < CmdMaxArgs {
			e.args = append(e.args, token)
		}
	}
}

func (e *CmdExecutor) ExecuteString(line string, source CmdSource) bool {
	e.source = source
	e.TokenizeBuffer([]rune(line))

	if len(e.args) == 0 {
		return true
	}

	for cmd := e.functions; cmd != nil; cmd = cmd.Next {
		if cmd.Name == e.args[0] {
			if source == CmdSourceClient && cmd.Source != CmdSourceClient {
				// TODO: Report client
				// log.Printf("%s tried to %s\n", )
				return false
			} else if source == CmdSourceCommand && cmd.Source == CmdSourceServer {
				continue
			} else if source == CmdSourceServer && cmd.Source != CmdSourceServer {
				continue
			}

			cmd.Function()
			return true
		}
	}

	if source == CmdSourceClient {
		// TODO: Report client
		return false
	}

	if source != CmdSourceCommand {
		return false
	}

	for a := e.aliases; a != nil; a = a.Next {
		if a.Name == e.args[0] {
			e.InsertText(a.Value)
			return true
		}
	}

	if !CVars.HandleCVarCommand() {
		if CVarClWarncmd.Value != 0 { // TODO: developer.value
			log.Printf("Unknown command: \"%s\"\n", e.args[0])
		}
	}

	return true
}

func (e *CmdExecutor) TintSubstring(value string, substr string) string {
	tintedRunes := []rune(substr)
	for runeIndex, r := range tintedRunes {
		tintedRunes[runeIndex] = r | 0x80
	}
	tintedSubstr := string(tintedRunes)

	splits := strings.Split(value, substr)
	var output strings.Builder

	if len(splits) > 0 {
		output.WriteString(splits[0])
	}

	for i := 1; i < len(splits); i++ {
		output.WriteString(tintedSubstr)
		output.WriteString(splits[i])
	}

	return output.String()
}

func (e *CmdExecutor) CmdList() {
	var partial string

	if len(e.args) > 1 {
		partial = e.args[1]
	} else {
		partial = ""
	}

	var count int
	for cmd := e.functions; cmd != nil; cmd = cmd.Next {
		if partial != "" && !strings.HasPrefix(cmd.Name, partial) {
			continue
		}

		log.Printf("   %s\n", cmd.Name)
		count++
	}

	log.Printf("%d commands", count)
	if partial != "" {
		log.Printf(" beginning with \"%s\"", partial)
	}
	log.Println()
}

func (e *CmdExecutor) CmdUnalias() {
	switch len(e.args) {
	default:
		log.Println("unalias <name> : delete alias")
		break
	case 2:
		var prev *CmdAlias
		for a := e.aliases; a != nil; a = a.Next {
			if e.args[1] == a.Name {
				if prev != nil {
					prev.Next = a.Next
				} else {
					e.aliases = a.Next
				}

				return
			}

			prev = a
		}

		log.Printf("No alias named %s\n", e.args[1])
		break
	}
}

func (e *CmdExecutor) CmdUnaliasAll() {
	e.aliases = nil
}

func (e *CmdExecutor) CmdExec() {
	if len(e.args) != 2 {
		log.Println("exec <filename> : execute a script file")
		return
	}

	var scriptBytes []byte
	var err error
	var scriptLoaded bool

	if MultiUser {
		scriptPath := path.Join(sdl.GetPrefPath("", "vkngQuake"), e.args[1])
		scriptBytes, err = os.ReadFile(scriptPath)
		if err == nil {
			scriptLoaded = true
		}
	}

	if !scriptLoaded {
		scriptBytes, _ = Files.LoadFile(e.args[1])
		if scriptBytes == nil && CVarClWarncmd.Value != 0 {
			log.Printf("couldn't exec %s\n", e.args[1])
		}
		return
	}
	if CVarClWarncmd.Value != 0 {
		log.Printf("execing %s\n", e.args[1])
	}

	e.InsertText(string(scriptBytes) + "\n")
}

func (e *CmdExecutor) CmdStuffCmds() {
	var j int
	var cmds [MaxCmdLineLength]rune
	cmdLine := CVarCmdline.StringVal
	plus := false

	for i := 0; i < len(cmdLine); i++ {
		if cmdLine[i] == '+' {
			plus = true
			if j > 0 {
				cmds[j-1] = ';'
				cmds[j] = ' '
				j++
			}
		} else if cmdLine[i] == '-' && (i == 0 || cmdLine[i-1] == ' ') {
			plus = false
		} else if plus {
			cmds[j] = rune(cmdLine[i])
			j++
		}
	}

	e.InsertText(string(cmds[:j]))
}

func (e *CmdExecutor) CmdAlias() {
	switch Cmds.ArgCount() {
	case 1:
		// List all aliases
		var i int
		for a := e.aliases; a != nil; a, i = a.Next, i+1 {
			log.Printf("   %s: %s", a.Name, a.Value)
		}
		if i > 0 {
			log.Printf("%d alias command(s)\n", i)
		} else {
			log.Println("no alias commands found")
		}
		break
	case 2:
		// Output current alias string
		for a := e.aliases; a != nil; a = a.Next {
			if Cmds.Arg(1) == a.Name {
				log.Printf("   %s: %s", a.Name, a.Value)
			}
		}

		break
	default:
		// Set alias string
		name := Cmds.Arg(1)
		if len(name) >= AliasMaxNameLength {
			log.Println("Alias name is too long")
			return
		}

		var newAlias *CmdAlias
		for newAlias = e.aliases; newAlias != nil; newAlias = newAlias.Next {
			if newAlias.Name == name {
				break
			}
		}

		if newAlias == nil {
			newAlias = &CmdAlias{
				Next: e.aliases,
			}
			e.aliases = newAlias
		}

		newAlias.Name = name
		var value strings.Builder

		for i := 2; i < Cmds.ArgCount(); i++ {
			value.WriteString(Cmds.Arg(i))
			if i != Cmds.ArgCount()-1 {
				value.WriteRune(' ')
			}
		}

		value.WriteRune('\n')
		if value.Len() >= 1024 {
			log.Println("alias value too long!")
			value.Reset()
			value.WriteRune('\n')
		}

		newAlias.Value = value.String()
		break
	}
}

func (e *CmdExecutor) CmdApropos() {
	substr := Cmds.Arg(1)
	var hits int
	if substr == "" {
		log.Printf("%s <substring> : search through commands and cvars for the given substring\n", Cmds.Arg(0))
		return
	}

	for cmd := e.functions; cmd != nil; cmd = cmd.Next {
		if strings.HasPrefix(cmd.Name, substr) && cmd.Source != CmdSourceServer {
			hits++
			log.Printf("%s\n", e.TintSubstring(cmd.Name, substr))
		}
	}

	for cvar := CVars.FindVarAfter("", 0); cvar != nil; cvar = cvar.Next {
		if strings.HasPrefix(cvar.Name, substr) {
			hits++
			log.Printf("%s (current value \"%s\")\n", e.TintSubstring(cvar.Name, substr), cvar.StringVal)
		}
	}

	if hits == 0 {
		log.Println("no cvars nor commands contain that substring")
	}
}

func CmdEcho() {
	for i := 1; i < Cmds.ArgCount(); i++ {
		log.Printf("%s ", Cmds.Arg(i))
	}
	log.Println()
}
