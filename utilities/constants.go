package utilities

type Consts struct {
	EnvPrefix   string
	RootCmdName string
}

var Constants = Consts{
	EnvPrefix:   "FOLDCLI_",
	RootCmdName: "foldcli",
}
