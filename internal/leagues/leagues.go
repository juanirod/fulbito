package leagues

type League struct {
	Slug string
	Name string
	Cmd  string
}

var All = []League{
	{Slug: "arg.1", Name: "Liga Profesional Argentina", Cmd: "liga-arg"},
	{Slug: "uefa.champions", Name: "UEFA Champions League", Cmd: "champions"},
	{Slug: "uefa.europa", Name: "UEFA Europa League", Cmd: "europa-league"},
	{Slug: "conmebol.libertadores", Name: "Copa Libertadores", Cmd: "libertadores"},
	{Slug: "conmebol.sudamericana", Name: "Copa Sudamericana", Cmd: "sudamericana"},
	{Slug: "eng.1", Name: "Premier League", Cmd: "premier"},
	{Slug: "esp.1", Name: "La Liga", Cmd: "laliga"},
	{Slug: "ita.1", Name: "Serie A", Cmd: "seriea"},
	{Slug: "ger.1", Name: "Bundesliga", Cmd: "bundesliga"},
	{Slug: "fra.1", Name: "Ligue 1", Cmd: "ligue1"},
	{Slug: "usa.1", Name: "MLS", Cmd: "mls"},
}
