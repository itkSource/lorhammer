package model

//CommandName is a type to be sure command are send (and not all string possible)
type CommandName string

//All commands for communication between orchestrator and lorhammer
const (
	INIT     = "init"
	REGISTER = "register"
	START    = "start"
	STOP     = "stop"
	SHUTDOWN = "shutdown"
)
