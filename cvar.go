package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
)

type CVarFlags int

const (
	CVarFlagArchive = 1 << iota
	CVarFlagNotify
	CVarFlagServerInfo
	CVarFlagUserInfo
	CVarFlagChanged
	CVarFlagROM
	CVarFlagLocked
	CVarFlagRegistered
	CVarFlagCallback
	CVarFlagUserDefined
	CVarFlagAutoCVar
	CVarFlagSeta
)

var CVars *CVarLibrary = &CVarLibrary{}

type CVarCallbackFunc func(cvar *CVar)

type CVar struct {
	Name          string
	StringVal     string
	Value         float64
	Flags         CVarFlags
	DefaultString string
	Callback      CVarCallbackFunc
	Next          *CVar
}

type CVarLibrary struct {
	vars *CVar
}

func (l *CVarLibrary) Init() {
	Cmds.Add("cvarlist", l.CmdList, CmdSourceCommand)
	Cmds.Add("toggle", l.CmdToggle, CmdSourceCommand)
	Cmds.Add("cycle", l.CmdCycle, CmdSourceCommand)
	Cmds.Add("inc", l.CmdInc, CmdSourceCommand)
	Cmds.Add("reset", l.CmdReset, CmdSourceCommand)
	Cmds.Add("resetall", l.CmdResetAll, CmdSourceCommand)
	Cmds.Add("resetcfg", l.CmdResetCfg, CmdSourceCommand)
	Cmds.Add("set", l.CmdSet, CmdSourceCommand)
	Cmds.Add("seta", l.CmdSet, CmdSourceCommand)
}

func (l *CVarLibrary) FindVar(name string) *CVar {
	for v := l.vars; v != nil; v = v.Next {
		if v.Name == name {
			return v
		}
	}

	return nil
}

func (l *CVarLibrary) FindVarAfter(prevName string, withFlags CVarFlags) *CVar {
	var v *CVar
	if prevName == "" {
		v = l.vars
	} else {
		prev := l.FindVar(prevName)
		if prev == nil {
			return nil
		}
		v = prev.Next
	}

	for v != nil {
		if withFlags == 0 || (v.Flags&withFlags) == withFlags {
			break
		}

		v = v.Next
	}

	return v
}

func (l *CVarLibrary) Lock(varName string) {
	v := l.FindVar(varName)
	if v != nil {
		v.Flags |= CVarFlagLocked
	}
}

func (l *CVarLibrary) Unlock(varName string) {
	v := l.FindVar(varName)
	if v != nil {
		v.Flags &= ^CVarFlagLocked
	}
}

func (l *CVarLibrary) UnlockAll() {
	for v := l.vars; v != nil; v = v.Next {
		v.Flags &= ^CVarFlagLocked
	}
}

func (l *CVarLibrary) Value(varName string) float64 {
	v := l.FindVar(varName)
	if v == nil {
		return 0
	}

	return v.Value
}

func (l *CVarLibrary) String(varName string) string {
	v := l.FindVar(varName)
	if v == nil {
		return ""
	}

	return v.StringVal
}

func (l *CVarLibrary) CompleteVariableName(partialName string) string {
	if partialName == "" {
		return ""
	}

	for v := l.vars; v != nil; v = v.Next {
		if strings.HasPrefix(v.Name, partialName) {
			return v.Name
		}
	}

	return ""
}

func (l *CVarLibrary) Reset(varName string) {
	v := l.FindVar(varName)
	if v == nil {
		log.Printf("variable \"%s\" not found\n", varName)
	} else {
		l.SetQuick(v, v.DefaultString)
	}
}

func (l *CVarLibrary) SetQuick(v *CVar, value string) {
	if v.Flags&(CVarFlagROM|CVarFlagLocked) != 0 {
		return
	}
	if v.Flags&CVarFlagRegistered == 0 {
		return
	}

	if v.StringVal != "" && v.StringVal == value {
		// no change
		return
	}

	if v.StringVal != "" {
		v.Flags |= CVarFlagChanged
	}
	v.StringVal = value

	numVal, err := strconv.ParseFloat(value, 64)
	if err == nil {
		v.Value = numVal
	} else {
		v.Value = 0
	}

	if v.DefaultString == "" || !HostInitialized {
		v.DefaultString = value
	}

	if v.Callback != nil {
		v.Callback(v)
	}
	if v.Flags&CVarFlagAutoCVar != 0 {
		//TODO: PR_AutoCvarChanged(v)
	}
}

func (l *CVarLibrary) SetValueQuick(v *CVar, value float64) {
	intVal := int64(value + 0.5)

	var strVal string
	if math.Abs(value-float64(intVal)) < 0.0001 {
		strVal = strconv.FormatInt(intVal, 10)
	} else {
		strVal = strconv.FormatFloat(value, 'f', -1, 64)
	}

	l.SetQuick(v, strVal)
}

func (l *CVarLibrary) Set(varName string, value string) {
	v := l.FindVar(varName)
	if v == nil {
		log.Printf("Cvar_Set: variable %s not found\n", varName)
		return
	}

	l.SetQuick(v, value)
}

func (l *CVarLibrary) SetValue(varName string, value float64) {
	v := l.FindVar(varName)
	if v == nil {
		log.Printf("Cvar_Set: variable %s not found\n", varName)
		return
	}

	l.SetValueQuick(v, value)
}

func (l *CVarLibrary) SetROM(varName string, value string) {
	v := l.FindVar(varName)
	if v != nil {
		v.Flags &= ^CVarFlagROM
		l.SetQuick(v, value)
		v.Flags |= CVarFlagROM
	}
}

func (l *CVarLibrary) SetValueROM(varName string, value float64) {
	v := l.FindVar(varName)
	if v != nil {
		v.Flags &= ^CVarFlagROM
		l.SetValueQuick(v, value)
		v.Flags |= CVarFlagROM
	}
}

func (l *CVarLibrary) Register(variable *CVar) {
	existingVar := l.FindVar(variable.Name)
	if existingVar != nil {
		log.Printf("Can't register variable %s, already defined\n", variable.Name)
		return
	}

	if Cmds.Exists(variable.Name) {
		log.Printf("Cvar_RegisterVariable: %s is a command\n", variable.Name)
		return
	}

	if l.vars == nil || strings.Compare(variable.Name, l.vars.Name) < 0 {
		l.vars.Next = variable
		l.vars = variable
	} else {
		prev := l.vars
		cursor := l.vars.Next

		for cursor != nil && strings.Compare(variable.Name, cursor.Name) > 0 {
			prev = cursor
			cursor = cursor.Next
		}

		variable.Next = prev.Next
		prev.Next = variable
	}

	value := variable.StringVal
	variable.Flags |= CVarFlagRegistered
	variable.Value = 0
	variable.StringVal = ""
	variable.DefaultString = ""

	if variable.Flags&CVarFlagCallback == 0 {
		variable.Callback = nil
	}

	isROM := variable.Flags & CVarFlagROM
	variable.Flags &= ^CVarFlagROM
	l.SetQuick(variable, value)
	variable.Flags |= isROM
}

func (l *CVarLibrary) Create(varName string, value string) *CVar {
	v := l.FindVar(varName)

	if v != nil {
		return v
	}
	if Cmds.Exists(varName) {
		return nil
	}

	v = &CVar{
		Name:      varName,
		Flags:     CVarFlagUserDefined,
		StringVal: value,
	}
	l.Register(v)
	return v
}

func (l *CVarLibrary) SetCallback(v *CVar, callback CVarCallbackFunc) {
	v.Callback = callback
	if v.Callback != nil {
		v.Flags |= CVarFlagCallback
	} else {
		v.Flags &= ^CVarFlagCallback
	}
}

func (l *CVarLibrary) HandleCVarCommand() bool {
	argCount := Cmds.ArgCount()

	if argCount == 0 {
		return false
	}

	v := l.FindVar(Cmds.Arg(0))
	if v == nil {
		return false
	}

	if argCount == 1 {
		log.Printf("\"%s\" is \"%s\"\n", v.Name, v.StringVal)
		return true
	}

	l.Set(Cmds.Arg(0), Cmds.Arg(1))
	return true
}

func (l *CVarLibrary) WriteVariables(writer io.Writer) error {
	var err error
	for v := l.vars; v != nil; v = v.Next {
		if v.Flags&CVarFlagArchive != 0 {
			if v.Flags&(CVarFlagUserDefined|CVarFlagSeta) != 0 {
				_, err = fmt.Fprint(writer, "seta ")
				if err != nil {
					return err
				}

				_, err = fmt.Fprintf(writer, "%s \"%s\"\n", v.Name, v.StringVal)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (l *CVarLibrary) CmdList() {
	var partial string

	if Cmds.ArgCount() > 1 {
		partial = Cmds.Arg(1)
	}

	var count int
	for cvar := l.vars; cvar != nil; cvar = cvar.Next {
		if partial != "" && !strings.HasPrefix(cvar.Name, partial) {
			continue
		}

		archiveIndicator := " "
		notifyIndicator := " "
		if cvar.Flags&CVarFlagArchive != 0 {
			archiveIndicator = "*"
		}
		if cvar.Flags&CVarFlagNotify != 0 {
			notifyIndicator = "s"
		}

		log.Printf("%s%s %s \"%s\"\n", archiveIndicator, notifyIndicator, cvar.Name, cvar.StringVal)
		count++
	}

	log.Printf("%d cvars", count)
	if partial != "" {
		log.Printf(" beginning with \"%s\"", partial)
	}
	log.Println()
}

func (l *CVarLibrary) CmdToggle() {
	if Cmds.ArgCount() < 2 {
		log.Printf("toggle <cvar> [value] [altvalue]: toggle cvar\n")
		return
	}

	cvar := l.FindVar(Cmds.Arg(1))
	if cvar == nil {
		log.Printf("variable \"%s\" not found\n", Cmds.Arg(1))
		return
	}

	if Cmds.ArgCount() >= 3 {
		newVal := Cmds.Arg(2)
		defaultVal := cvar.DefaultString

		if Cmds.ArgCount() > 3 {
			defaultVal = Cmds.Arg(3)
		}

		if cvar.StringVal == newVal {
			l.SetQuick(cvar, defaultVal)
		} else {
			l.SetQuick(cvar, newVal)
		}
		return
	}

	if cvar.Value == 0 {
		l.SetQuick(cvar, "1")
	} else {
		l.SetQuick(cvar, "0")
	}
}

func (l *CVarLibrary) CmdInc() {
	varName := Cmds.Arg(1)

	switch Cmds.ArgCount() {
	default:
		fallthrough
	case 1:
		log.Println("inc <cvar> [amount] : increment cvar")
		break
	case 2:
		l.SetValue(varName, l.Value(varName)+1)
		break
	case 3:
		val, err := strconv.ParseFloat(Cmds.Arg(2), 64)
		if err != nil {
			val = 1
		}
		l.SetValue(varName, l.Value(varName)+val)
		break
	}
}

func (l *CVarLibrary) CmdCycle() {
	if Cmds.ArgCount() < 3 {
		log.Println("cycle <cvar> <value list>: cycle cvar through a list of values")
		return
	}

	cvar := l.FindVar(Cmds.Arg(1))
	if cvar == nil {
		log.Printf("variable \"%s\" not found\n", Cmds.Arg(1))
		return
	}

	// loop through the args until you find one that matches the current cvar value.
	// yes, this will get stuck on a list that contains the same value twice
	matchIndex := Cmds.ArgCount() - 1
	for i := 2; i < Cmds.ArgCount()-1; i++ {
		argValue := Cmds.Arg(i)
		argFloatValue, err := strconv.ParseFloat(argValue, 64)
		if err != nil && argValue == cvar.StringVal {
			// This is a string
			matchIndex = i
			break
		} else if err == nil && math.Abs(argFloatValue-cvar.Value) < 0.0001 {
			// This is a float
			matchIndex = i
			break
		}
	}

	matchIndex++
	if matchIndex == Cmds.ArgCount() {
		matchIndex = 2
	}

	l.Set(cvar.Name, Cmds.Arg(matchIndex))
}

func (l *CVarLibrary) CmdReset() {
	switch Cmds.ArgCount() {
	default:
		fallthrough
	case 1:
		log.Println("reset <cvar> : reset cvar to default")
		break
	case 2:
		l.Reset(Cmds.Arg(1))
		break
	}
}

func (l *CVarLibrary) CmdResetAll() {
	for cvar := l.vars; cvar != nil; cvar = cvar.Next {
		l.Reset(cvar.Name)
	}
}

func (l *CVarLibrary) CmdResetCfg() {
	for cvar := l.vars; cvar != nil; cvar = cvar.Next {
		if cvar.Flags&CVarFlagArchive != 0 {
			l.Reset(cvar.Name)
		}
	}
}

func (l *CVarLibrary) CmdSet() {
	varName := Cmds.Arg(1)
	varValue := Cmds.Arg(2)

	if Cmds.ArgCount() < 3 {
		log.Printf("%s <cvar> <value>\n", Cmds.Arg(0))
		return
	}

	if Cmds.ArgCount() > 3 {
		// TODO: Add warning log
		log.Printf("%s \"%s\" command with extra args\n", Cmds.Arg(0), varName)
		return
	}

	cvar := l.Create(varName, varValue)
	if cvar == nil {
		log.Printf("%s could not create a cvar named \"%s\"", Cmds.Arg(0), varName)
		return
	}

	l.SetQuick(cvar, varValue)

	if Cmds.Arg(0) == "seta" {
		cvar.Flags |= CVarFlagArchive | CVarFlagSeta
	}
}
