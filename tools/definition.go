package tools

// Tool représente un outil individuel
type Tool struct {
	Name        string
	Description string
	Action      func() // La fonction à lancer quand on clique dessus
}

// Category regroupe des outils
type Category struct {
	Name  string
	Tools []Tool
}

// MockData génère des fausses données pour tester l'UI
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
			Name: "Système",
			Tools: []Tool{
				{Name: "Info CPU", Description: "Affiche l'usage CPU"},
				{Name: "Disk Usage", Description: "Espace disque"},
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
