package creatures

type Boss struct {
	Entry              int
	Name               string
	ScriptName         string `db:"ScriptName"`
	ExperienceModifier int    `db:"ExperienceModifier"`
}
