package model

//CommandName is a type to be sure command are send (and not all string possible)
type CommandName string

//All commands for communication between orchestrator and lorhammer
const (
	NEWLORHAMMER   = "newLorhammer"   // prevent orchestrator new lorhammer is here LORHAMMER -> ORCHESTRATOR
	LORHAMMERADDED = "lorhammerAdded" // orchestrator has received lorhammer infos ORCHESTRATOR -> LORHAMME
	INIT           = "init"           // send init (lorawan info) ORCHESTRATOR -> LORHAMMER
	REGISTER       = "register"       // send lorawan sensors to provision LORHAMMER -> ORCHESTRATOR
	START          = "start"          // send start after sensors provisioning ORCHESTRATOR -> LORHAMMER
	STOP           = "stop"           // send stop to finish test ORCHESTRATOR -> LORHAMMER
	SHUTDOWN       = "shutdown"       // kill lorhammers ORCHESTRATOR -> LORHAMMER
)

//NewLorhammer is the struct used by lorhammer to prevent orchestrator
type NewLorhammer struct {
	CallbackTopic string
}
