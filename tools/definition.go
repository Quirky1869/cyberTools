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
			Name: "Réseau",
			Tools: []Tool{
				{Name: "Ping Test", Description: "Ping une IP"},
				{Name: "Port Scan", Description: "Scan les ports ouverts"},
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
			Name: "Crypto",
			Tools: []Tool{
				{Name: "Hash Generator", Description: "MD5/SHA256"},
			},
		},
	}
}