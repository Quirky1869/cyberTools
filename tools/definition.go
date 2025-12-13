package tools

type Tool struct {
	Name        string
	Description string
	Action      func()
}

type Category struct {
	Name  string
	Tools []Tool
}

func GetCategories() []Category {
	return []Category{
		{
			Name: "BDD",
			Tools: []Tool{
				{Name: "SqlTUI", Description: "Explorateur SQL"},
			},
		},
		{
			Name: "Utilitaire",
			Tools: []Tool{
				{Name: "LogV", Description: "Visualiseur de logs"},
				{Name: "structViewer", Description: "Lecteur YAML/JSON arborescent"},
			},
		},
		{
			Name: "Data",
			Tools: []Tool{
				{Name: "AED", Description: "Analyseur d'espace disque"},
			},
		},
	}
}
