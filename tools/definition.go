package tools

// Tool représente un outil individuel
type Tool struct {
	Name        string
	Description string
	Action      func() // Sera géré dans root.go
}

// Category regroupe des outils
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
