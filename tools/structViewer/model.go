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

// Signal de retour au menu principal
type BackMsg struct{}

// États de la machine à états (sélection ou navigation)
type SessionState int

const (
	StateSelectFile SessionState = iota
	StateTree
)

// Structure récursive représentant un nœud de l'arbre JSON/YAML
type Node struct {
	Key      string
	Value    string
	Level    int
	Children []*Node
	Expanded bool
	IsLeaf   bool
}

// Définition des styles de l'interface
var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true)
	keyStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true)
	valStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))
	pathStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#500aff")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Bold(true)
)

// Modèle principal contenant l'état du sélecteur de fichier et de l'arbre de données
type Model struct {
	state      SessionState
	filePicker filepicker.Model

	root      *Node
	flatNodes []*Node
	cursor    int
	yOffset   int

	width, height int
	err           error
}

// Initialisation du modèle et configuration du FilePicker avec le thème
func New(w, h int) Model {
	fp := filepicker.New()

	fp.AllowedTypes = []string{".yaml", ".yml", ".json", ".txt", ".md", ".go", ".mod"}
	fp.CurrentDirectory, _ = os.Getwd()

	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D"))
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true)
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))
	fp.Styles.File = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	fpHeight := h - 8
	if fpHeight < 5 {
		fpHeight = 5
	}
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

		// Gestion de la vue sélection de fichier
		if m.state == StateSelectFile {
			switch msg.String() {
			case "q", "esc":
				return m, func() tea.Msg { return BackMsg{} }
			case "h":
				m.filePicker.ShowHidden = !m.filePicker.ShowHidden
				return m, m.filePicker.Init()
			}
		} else if m.state == StateTree {
			// Gestion de la vue Arbre (Tree View)
			switch msg.String() {
			case "q", "esc":
				m.state = StateSelectFile
				m.root = nil
				m.flatNodes = nil
				m.cursor = 0
				m.yOffset = 0
				return m, m.filePicker.Init()

			// Navigation verticale avec gestion du scroll
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					if m.cursor < m.yOffset {
						m.yOffset = m.cursor
					}
				}
			case "down", "j":
				if m.cursor < len(m.flatNodes)-1 {
					m.cursor++
					visibleHeight := m.height - 4
					if m.cursor >= m.yOffset+visibleHeight {
						m.yOffset = m.cursor - visibleHeight + 1
					}
				}

			// Expansion / Réduction des nœuds
			case "enter", " ", "space", "right", "l":
				node := m.flatNodes[m.cursor]
				if !node.IsLeaf {
					node.Expanded = !node.Expanded
					m.updateFlatNodes()
				}
			case "left", "h":
				node := m.flatNodes[m.cursor]
				if node.Expanded {
					node.Expanded = false
					m.updateFlatNodes()
				}
			}
			return m, nil
		}
	}

	// Mise à jour du composant enfant FilePicker
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
	// Vue 1 : Sélection du fichier
	if m.state == StateSelectFile {
		title := titleStyle.Render("structViewer - Ouvrir un fichier")
		currentDir := fmt.Sprintf(" %s", pathStyle.Render(m.filePicker.CurrentDirectory))
		content := "\n" + m.filePicker.View()

		hiddenStatus := "off"
		if m.filePicker.ShowHidden {
			hiddenStatus = "ON"
		}
		helpText := fmt.Sprintf("↑/↓/←/→: naviguer • enter: ouvrir • h: cachés(%s) • q: retour", hiddenStatus)
		footer := helpStyle.Render("\n" + helpText)

		return fmt.Sprintf("\n  %s\n\n  %s%s%s", title, currentDir, content, footer)
	}

	// Vue 2 : Affichage de l'arbre JSON/YAML
	if m.state == StateTree {
		if m.err != nil {
			return fmt.Sprintf("Erreur: %v\n(q pour quitter)", m.err)
		}
		if m.root == nil {
			return "Fichier vide ou invalide."
		}

		header := titleStyle.Render(" Navigateur YAML/JSON - structViewer \n")

		var lines []string

		// Calcul de la zone visible (Virtual Scrolling)
		visibleHeight := m.height - 5
		if visibleHeight < 1 {
			visibleHeight = 1
		}

		end := m.yOffset + visibleHeight
		if end > len(m.flatNodes) {
			end = len(m.flatNodes)
		}

		// Rendu des lignes visibles
		for i := m.yOffset; i < end; i++ {
			node := m.flatNodes[i]

			indent := strings.Repeat("  ", node.Level)

			icon := "  "
			if !node.IsLeaf {
				if node.Expanded {
					icon = "▼ "
				} else {
					icon = "▶ "
				}
			}

			lineContent := ""
			if node.Key != "" {
				lineContent += keyStyle.Render(node.Key + ": ")
			}
			if node.Value != "" {
				lineContent += valStyle.Render(node.Value)
			}

			fullLine := fmt.Sprintf("%s%s%s", indent, icon, lineContent)

			if i == m.cursor {
				fullLine = selectedStyle.Render(fmt.Sprintf("%-*s", m.width-2, fullLine))
			}

			lines = append(lines, fullLine)
		}

		footer := helpStyle.Render("\n↑/↓: scroll • enter/space: ouvrir/fermer • q: retour")

		return header + "\n" + strings.Join(lines, "\n") + "\n" + footer
	}

	return "Loading..."
}

// Charge le fichier, parse le YAML/JSON et construit l'arbre initial
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

	m.root = buildTree("root", data, 0)
	m.root.Expanded = true

	m.updateFlatNodes()
	return nil
}

// Met à jour la liste plate (flatNodes) utilisée pour l'affichage en parcourant récursivement les nœuds ouverts
func (m *Model) updateFlatNodes() {
	m.flatNodes = []*Node{}
	if m.root != nil {
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

// Construit récursivement la structure de nœuds à partir des données brutes (Map ou Slice)
func buildTree(key string, data interface{}, level int) *Node {
	node := &Node{
		Key:   key,
		Level: level,
	}

	switch v := data.(type) {
	case map[string]interface{}:
		node.IsLeaf = false
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
			child := buildTree(fmt.Sprintf("- [%d]", i), item, level+1)
			node.Children = append(node.Children, child)
		}

	default:
		node.IsLeaf = true
		node.Value = fmt.Sprintf("%v", v)
	}

	return node
}