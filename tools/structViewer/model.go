package structviewer

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// --- TYPES & MESSAGES ---

type BackMsg struct{}

type SessionState int

const (
	StateSelectFile SessionState = iota
	StateTree
)

// Node représente un élément de l'arbre YAML
type Node struct {
	Key      string
	Value    string
	Level    int
	Children []*Node
	Expanded bool
	IsLeaf   bool
}

// --- STYLES ---

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true)
	keyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true) // Pink
	valStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))            // Cyan

	// AJOUTS UX
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))
	pathStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#500aff")).Bold(true)

	// Curseur
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Bold(true)
)

// --- MODEL ---

type Model struct {
	state      SessionState
	filePicker filepicker.Model

	// Données de l'arbre
	root      *Node
	flatNodes []*Node // Liste linéaire des nœuds visibles (pour l'affichage)
	cursor    int     // Position du curseur dans flatNodes
	yOffset   int     // Scroll vertical

	width, height int
	err           error
}

func New(w, h int) Model {
	fp := filepicker.New()
	
	fp.AllowedTypes = []string{".yaml", ".yml", ".json", ".txt", ".md", ".go", ".mod"}
	fp.CurrentDirectory, _ = os.Getwd()
	
	// --- PERSONNALISATION DES COULEURS ---
	
	// 1. Couleur du curseur (la petite flèche >)
	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")) // Rose
	
	// 2. Couleur du texte sélectionné
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true) // Rose Gras
	
	// 3. (Optionnel) Couleur des dossiers
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff")) // Cyan
	
	// 4. (Optionnel) Couleur des fichiers normaux
	fp.Styles.File = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")) // Blanc

	// -------------------------------------

	fpHeight := h - 8
	if fpHeight < 5 { fpHeight = 5 }
	fp.Height = fpHeight
	fp.ShowHidden = false

	return Model{
		state:      StateSelectFile,
		filePicker: fp,
		width:      w,
		height:     h,
	}
}

func (m Model) Init() tea.Cmd {
	return m.filePicker.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.filePicker.Height = msg.Height - 8

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, func() tea.Msg { return BackMsg{} }
		}

		// Gestion spécifique selon l'état
		if m.state == StateSelectFile {
			switch msg.String() {
			case "q", "esc":
				return m, func() tea.Msg { return BackMsg{} }
			case "h":
				// AJOUT UX: Toggle fichiers cachés
				m.filePicker.ShowHidden = !m.filePicker.ShowHidden
				return m, m.filePicker.Init()
			}
		} else if m.state == StateTree {
			switch msg.String() {
			case "q", "esc":
				// Retour au file picker
				m.state = StateSelectFile
				m.root = nil
				m.flatNodes = nil
				m.cursor = 0
				m.yOffset = 0
				return m, m.filePicker.Init()

			// Navigation
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					// Scroll up si besoin
					if m.cursor < m.yOffset {
						m.yOffset = m.cursor
					}
				}
			case "down", "j":
				if m.cursor < len(m.flatNodes)-1 {
					m.cursor++
					// Scroll down si besoin
					// m.height - 4 (header/footer)
					visibleHeight := m.height - 4
					if m.cursor >= m.yOffset+visibleHeight {
						m.yOffset = m.cursor - visibleHeight + 1
					}
				}

			// Action
			case "enter", " ", "space", "right", "l":
				node := m.flatNodes[m.cursor]
				if !node.IsLeaf {
					node.Expanded = !node.Expanded
					// Si on ferme, on doit s'assurer que si on était dans un enfant, on remonte
					m.updateFlatNodes()
				}
			case "left", "h":
				node := m.flatNodes[m.cursor]
				if node.Expanded {
					node.Expanded = false
					m.updateFlatNodes()
				} else {
					// Si déjà fermé, on pourrait remonter au parent (logique plus complexe, optionnel)
				}
			}
			return m, nil
		}
	}

	// Update des composants enfants
	if m.state == StateSelectFile {
		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)

		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			err := m.loadAndParseYAML(path)
			if err != nil {
				m.err = err
			} else {
				m.state = StateTree
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.state == StateSelectFile {
		// --- VUE AMELIOREE (Comme LogV) ---

		// 1. Titre
		title := titleStyle.Render("structViewer - Ouvrir un fichier")

		// 2. PWD
		currentDir := fmt.Sprintf(" %s", pathStyle.Render(m.filePicker.CurrentDirectory))

		// 3. Filepicker
		content := "\n" + m.filePicker.View()

		// 4. Footer Aide
		hiddenStatus := "off"
		if m.filePicker.ShowHidden {
			hiddenStatus = "ON"
		}
		helpText := fmt.Sprintf("↑/↓/←/→: naviguer • enter: ouvrir • h: cachés(%s) • q: retour", hiddenStatus)
		footer := helpStyle.Render("\n" + helpText)

		return fmt.Sprintf("\n  %s\n\n  %s%s%s", title, currentDir, content, footer)
	}

	if m.state == StateTree {
		if m.err != nil {
			return fmt.Sprintf("Erreur: %v\n(q pour quitter)", m.err)
		}
		if m.root == nil {
			return "Fichier vide ou invalide."
		}

		// Header
		header := titleStyle.Render(" Navigateur YAML/JSON - structViewer \n")

		// Contenu Tree View
		var lines []string

		// Calcul de la zone visible
		visibleHeight := m.height - 5 // un peu moins haut a cause du footer
		if visibleHeight < 1 {
			visibleHeight = 1
		}

		end := m.yOffset + visibleHeight
		if end > len(m.flatNodes) {
			end = len(m.flatNodes)
		}

		// Rendu des lignes visibles uniquement
		for i := m.yOffset; i < end; i++ {
			node := m.flatNodes[i]

			// Indentation
			indent := strings.Repeat("  ", node.Level)

			// Icône
			icon := "  " // Feuille
			if !node.IsLeaf {
				if node.Expanded {
					icon = "▼ "
				} else {
					icon = "▶ "
				}
			}

			// Texte
			lineContent := ""
			if node.Key != "" {
				lineContent += keyStyle.Render(node.Key + ": ")
			}
			if node.Value != "" {
				lineContent += valStyle.Render(node.Value)
			}

			fullLine := fmt.Sprintf("%s%s%s", indent, icon, lineContent)

			// Application du style de sélection
			if i == m.cursor {
				// On force la ligne à prendre toute la largeur pour voir la sélection
				fullLine = selectedStyle.Render(fmt.Sprintf("%-*s", m.width-2, fullLine))
			}

			lines = append(lines, fullLine)
		}

		// Footer pour l'arbre
		footer := helpStyle.Render("\n↑/↓: scroll • enter/space: ouvrir/fermer • q: retour")

		return header + "\n" + strings.Join(lines, "\n") + "\n" + footer
	}

	return "Loading..."
}

// --- LOGIQUE METIER ---

// loadAndParseYAML lit le fichier et construit l'arbre
func (m *Model) loadAndParseYAML(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var data interface{}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return err
	}

	// Construction de l'arbre racine
	m.root = buildTree("root", data, 0)
	m.root.Expanded = true // La racine est ouverte par défaut

	// Initialisation de la vue plate
	m.updateFlatNodes()
	return nil
}

// updateFlatNodes met à jour la liste linéaire utilisée par la View
// en parcourant l'arbre selon l'état Expanded de chaque nœud
func (m *Model) updateFlatNodes() {
	m.flatNodes = []*Node{}
	if m.root != nil {
		// On saute la racine "virtuelle" pour l'affichage, on affiche direct ses enfants
		// ou alors on affiche la racine si on veut. Ici on affiche les enfants direct pour faire propre.
		for _, child := range m.root.Children {
			m.recurseFlatten(child)
		}
	}
}

func (m *Model) recurseFlatten(node *Node) {
	m.flatNodes = append(m.flatNodes, node)
	if node.Expanded {
		for _, child := range node.Children {
			m.recurseFlatten(child)
		}
	}
}

// buildTree construit récursivement les nœuds
func buildTree(key string, data interface{}, level int) *Node {
	node := &Node{
		Key:   key,
		Level: level,
	}

	switch v := data.(type) {
	case map[string]interface{}:
		node.IsLeaf = false
		// Tri des clés pour un affichage stable
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			child := buildTree(k, v[k], level+1)
			node.Children = append(node.Children, child)
		}

	case []interface{}:
		node.IsLeaf = false
		for i, item := range v {
			// Pour les listes, la clé est l'index (ex: "- [0]")
			child := buildTree(fmt.Sprintf("- [%d]", i), item, level+1)
			node.Children = append(node.Children, child)
		}

	default:
		// C'est une feuille (string, int, bool...)
		node.IsLeaf = true
		node.Value = fmt.Sprintf("%v", v)
	}

	return node
}
