package typedefine

var systemData = &SystemData{}

type SystemData struct {
	Common *Common `json:"common" bson:"common"`
}

type Common struct {
	MaxGameServerId  int
	MaxCrossServerId int
}
