package aed

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Signal de retour au menu principal
type BackMsg struct{}

// Identifiant unique d'un fichier (device + inode) pour gérer les liens physiques
type fileID struct {
	dev uint64
	ino uint64
}

// Structure représentant un nœud dans l'arborescence de fichiers
type FileNode struct {
	Name     string
	Path     string
	Size     int64
	IsDir    bool
	Children []*FileNode
	Parent   *FileNode
}

// Machine à états pour gérer les différentes vues de l'outil
type SessionState int

const (
	StateInputPath SessionState = iota
	StateScanning
	StateBrowsing
)

// Message envoyé lorsque le scan récursif est terminé
type scanFinishedMsg struct {
	root *FileNode
	err  error
}

// Définition des styles visuels pour Lipgloss
var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true)
	pathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff")).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))

	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("#333333")).Bold(true)
	barFull       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D"))
	barEmpty      = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	countStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff")).Bold(true).PaddingLeft(2)
)

// Modèle principal contenant l'état du scanner et de l'interface
type Model struct {
	state     SessionState
	textInput textinput.Model
	spinner   spinner.Model

	root        *FileNode
	currentNode *FileNode
	cursor      int
	yOffset     int

	filesScanned *int64

	width, height int
	err           error
}

// Initialisation des composants (input, spinner) et des valeurs par défaut
func New(w, h int) Model {
	ti := textinput.New()
	ti.Placeholder = "/home/user"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.SetValue(".")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D"))

	zero := int64(0)

	return Model{
		state:        StateInputPath,
		textInput:    ti,
		spinner:      s,
		filesScanned: &zero,
		width:        w,
		height:       h,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Prépare la liste des éléments à afficher en ajoutant les entrées de navigation (. et ..)
func (m Model) getDisplayItems() []*FileNode {
	var items []*FileNode

	if m.currentNode == nil {
		return items
	}

	dot := &FileNode{
		Name:  ".",
		Path:  m.currentNode.Path,
		Size:  m.currentNode.Size,
		IsDir: true,
	}
	items = append(items, dot)

	if m.currentNode.Parent != nil {
		parentPath := filepath.Dir(m.currentNode.Path)
		dotdot := &FileNode{
			Name:  "..",
			Path:  parentPath,
			Size:  0,
			IsDir: true,
		}
		items = append(items, dotdot)
	}

	items = append(items, m.currentNode.Children...)

	return items
}

// Boucle principale de gestion des événements et des états
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, func() tea.Msg { return BackMsg{} }
		}

		// Gestion de la saisie du chemin à analyser
		if m.state == StateInputPath {
			switch msg.String() {
			case "enter":
				path := m.textInput.Value()
				if path == "." {
					path, _ = os.Getwd()
				}

				m.state = StateScanning
				atomic.StoreInt64(m.filesScanned, 0)
				visitedInodes := make(map[fileID]struct{})

				return m, tea.Batch(
					m.spinner.Tick,
					scanDirectoryCmd(path, m.filesScanned, visitedInodes),
				)
			case "esc", "q":
				return m, func() tea.Msg { return BackMsg{} }
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// Gestion pendant le scan (permet d'annuler)
		if m.state == StateScanning {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, func() tea.Msg { return BackMsg{} }
			}
		}

		// Gestion de la navigation dans les résultats
		if m.state == StateBrowsing {
			items := m.getDisplayItems()

			switch msg.String() {

			case "q":
				return m, func() tea.Msg { return BackMsg{} }

			// Ouvre le fichier ou dossier avec l'explorateur par défaut du système (xdg-open)
			case "g":
				if len(items) > 0 && m.cursor < len(items) {
					selected := items[m.cursor]
					cmd := exec.Command("xdg-open", selected.Path)
					cmd.Start()
				}
				return m, nil

			// Suspend l'interface pour ouvrir un shell dans le dossier sélectionné
			case "s":
				if len(items) > 0 && m.cursor < len(items) {
					selected := items[m.cursor]

					targetPath := selected.Path
					if !selected.IsDir {
						targetPath = filepath.Dir(selected.Path)
					}

					shell := os.Getenv("SHELL")
					if shell == "" {
						shell = "/bin/bash"
					}

					c := exec.Command(shell)
					c.Dir = targetPath

					return m, tea.ExecProcess(c, func(err error) tea.Msg {
						return nil
					})
				}
				return m, nil

			// Remonter au dossier parent
			case "backspace", "left", "h", "esc":
				if m.currentNode.Parent != nil {
					m.currentNode = m.currentNode.Parent
					m.cursor = 0
					m.yOffset = 0
				}

			// Entrer dans un dossier
			case "enter", "right", "l":
				if len(items) > 0 && m.cursor < len(items) {
					selected := items[m.cursor]
					if selected.Name == ".." {
						if m.currentNode.Parent != nil {
							m.currentNode = m.currentNode.Parent
							m.cursor = 0
							m.yOffset = 0
						}
						return m, nil
					}
					if selected.Name == "." {
						return m, nil
					}
					if selected.IsDir {
						for _, child := range m.currentNode.Children {
							if child.Path == selected.Path {
								m.currentNode = child
								m.cursor = 0
								m.yOffset = 0
								break
							}
						}
					}
				}

			// Déplacement du curseur
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					if m.cursor < m.yOffset {
						m.yOffset = m.cursor
					}
				}
			case "down", "j":
				if m.cursor < len(items)-1 {
					m.cursor++
					visibleHeight := m.height - 7
					if m.cursor >= m.yOffset+visibleHeight {
						m.yOffset = m.cursor - visibleHeight + 1
					}
				}
			}
		}

	// Réception du résultat du scan
	case scanFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = StateInputPath
		} else {
			m.root = msg.root
			m.currentNode = msg.root
			m.state = StateBrowsing
		}

	case spinner.TickMsg:
		if m.state == StateScanning {
			var cmdSpinner tea.Cmd
			m.spinner, cmdSpinner = m.spinner.Update(msg)
			return m, cmdSpinner
		}
	}

	return m, nil
}

func (m Model) View() string {
	// Vue 1 : Saisie du chemin
	if m.state == StateInputPath {
		title := titleStyle.Render("AED - Analyseur d'Espace Disque")
		input := m.textInput.View()
		return fmt.Sprintf("\n  %s\n\n  Entrez le dossier à analyser :\n  %s\n\n  %s", title, input, helpStyle.Render("(enter: valider • esc: quitter)"))
	}

	// Vue 2 : Chargement (Spinner)
	if m.state == StateScanning {
		count := atomic.LoadInt64(m.filesScanned)
		return fmt.Sprintf(
			"\n  %s Analyse en cours...\n\n%s fichiers scannés",
			m.spinner.View(),
			countStyle.Render(fmt.Sprintf("%d", count)),
		)
	}

	// Vue 3 : Explorateur de fichiers avec barres de taille
	if m.state == StateBrowsing {
		if m.currentNode == nil {
			return "Erreur: Node vide"
		}

		title := titleStyle.Render("AED")
		path := pathStyle.Render(m.currentNode.Path)
		totalSize := infoStyle.Render(fmt.Sprintf("Total: %s", formatBytes(m.currentNode.Size)))

		header := fmt.Sprintf("  %s  %s  (%s)\n", title, path, totalSize)

		var rows []string
		items := m.getDisplayItems()

		visibleHeight := m.height - 7
		if visibleHeight < 1 {
			visibleHeight = 1
		}

		start := m.yOffset
		end := start + visibleHeight
		if end > len(items) {
			end = len(items)
		}

		barWidth := 20

		for i := start; i < end; i++ {
			item := items[i]

			var sizeStr, bar, name string

			if item.Name == "." || item.Name == ".." {
				sizeStr = fmt.Sprintf("%8s", "")
				if item.Name == "." {
					sizeStr = fmt.Sprintf("%8s", formatBytes(item.Size))
				}
				bar = strings.Repeat(" ", barWidth)
				name = item.Name
			} else {
				sizeStr = fmt.Sprintf("%8s", formatBytes(item.Size))

				// Calcul du pourcentage pour la barre visuelle
				percent := 0.0
				if m.currentNode.Size > 0 {
					percent = float64(item.Size) / float64(m.currentNode.Size)
				}
				filledLen := int(percent * float64(barWidth))
				emptyLen := barWidth - filledLen
				bar = barFull.Render(strings.Repeat("■", filledLen)) + barEmpty.Render(strings.Repeat("-", emptyLen))

				name = item.Name
				if item.IsDir {
					name += "/"
				}
			}

			row := fmt.Sprintf("%s  %s  %s", sizeStr, bar, name)

			if i == m.cursor {
				row = selectedStyle.Render(fmt.Sprintf("%-*s", m.width-4, row))
			} else {
				row = fmt.Sprintf("  %s", row)
			}
			rows = append(rows, row)
		}

		content := strings.Join(rows, "\n")
		footer := helpStyle.Render("\n↑/↓/←/→: naviguer • enter: entrer • g: explorer • s: shell • q: quitter")

		return fmt.Sprintf("\n%s\n%s\n%s", header, content, footer)
	}

	return ""
}

// Commande Tea pour lancer le scan en arrière-plan
func scanDirectoryCmd(path string, counter *int64, visited map[fileID]struct{}) tea.Cmd {
	return func() tea.Msg {
		root, err := scanRecursively(path, nil, counter, visited)
		return scanFinishedMsg{root: root, err: err}
	}
}

// Fonction récursive principale : parcourt le disque, calcule les tailles et trie par taille décroissante
// Utilisation de 'visited' pour éviter de compter plusieurs fois les hardlinks
func scanRecursively(path string, parent *FileNode, counter *int64, visited map[fileID]struct{}) (*FileNode, error) {
	atomic.AddInt64(counter, 1)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(absPath)
	if parent == nil {
		name = absPath
	}

	node := &FileNode{
		Name:   name,
		Path:   absPath,
		IsDir:  true,
		Parent: parent,
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return node, nil
	}

	var totalSize int64

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Exclusion des pseudo-systèmes de fichiers sous Linux
		if node.Path == "/" && (entry.Name() == "proc" || entry.Name() == "sys" || entry.Name() == "dev" || entry.Name() == "run") {
			continue
		}

		childPath := filepath.Join(absPath, entry.Name())

		if entry.IsDir() {
			childNode, _ := scanRecursively(childPath, node, counter, visited)
			if childNode != nil {
				node.Children = append(node.Children, childNode)
				totalSize += childNode.Size
			}
		} else {
			atomic.AddInt64(counter, 1)

			var size int64
			// Calcul précis de la taille disque (blocks) et déduplication via inode/dev
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				size = stat.Blocks * 512
				id := fileID{dev: stat.Dev, ino: stat.Ino}
				if _, seen := visited[id]; !seen {
					visited[id] = struct{}{}
					totalSize += size
				}
			} else {
				size = info.Size()
				totalSize += size
			}

			child := &FileNode{
				Name:   entry.Name(),
				Path:   childPath,
				Size:   size,
				IsDir:  false,
				Parent: node,
			}
			node.Children = append(node.Children, child)
		}
	}

	node.Size = totalSize

	// Tri des enfants du plus gros au plus petit
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})

	return node, nil
}

// Formate les octets en unité lisible (KB, MB, GB...)
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
