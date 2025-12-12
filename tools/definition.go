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
				{Name: "Info CPU", Description: "Affiche l'usage CPU"},
			},
		},
		{
			Name: "Data",
			Tools: []Tool{
				{Name: "TreeYamlV", Description: ""},
			},
		},
	}
}

