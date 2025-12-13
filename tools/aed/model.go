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

// --- TYPES ---

type BackMsg struct{}

// Struct pour identifier un fichier de manière unique sur tout le système
type fileID struct {
	dev uint64
	ino uint64
}

type FileNode struct {
	Name     string
	Path     string
	Size     int64
	IsDir    bool
	Children []*FileNode
	Parent   *FileNode
}

type SessionState int

const (
	StateInputPath SessionState = iota
	StateScanning
	StateBrowsing
)

type scanFinishedMsg struct {
	root *FileNode
	err  error
}

// --- STYLES ---

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

// --- MODEL ---

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

// Helper pour générer la liste d'affichage avec "." et ".."
func (m Model) getDisplayItems() []*FileNode {
	var items []*FileNode

	if m.currentNode == nil {
		return items
	}

	// 1. Ajout de "."
	dot := &FileNode{
		Name:  ".",
		Path:  m.currentNode.Path,
		Size:  m.currentNode.Size,
		IsDir: true,
	}
	items = append(items, dot)

	// 2. Ajout de ".."
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

	// 3. Ajout des enfants
	items = append(items, m.currentNode.Children...)

	return items
}

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

		// ETAT 1 : INPUT
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

		// ETAT 2 : SCANNING
		if m.state == StateScanning {
			if msg.String() == "q" || msg.String() == "esc" {
				return m, func() tea.Msg { return BackMsg{} }
			}
		}

		// ETAT 3 : BROWSING
		if m.state == StateBrowsing {
			items := m.getDisplayItems()

			switch msg.String() {

			// --- QUITTER ---
			case "q":
				return m, func() tea.Msg { return BackMsg{} }

			// --- "g": EXPLORER (GUI) ---
			case "g":
				if len(items) > 0 && m.cursor < len(items) {
					selected := items[m.cursor]
					cmd := exec.Command("xdg-open", selected.Path)
					cmd.Start()
				}
				return m, nil

			// --- "s": OPEN SHELL (CLI) ---
			case "s":
				if len(items) > 0 && m.cursor < len(items) {
					selected := items[m.cursor]

					// On cible le dossier (si c'est un fichier, on prend le dossier parent)
					targetPath := selected.Path
					if !selected.IsDir {
						targetPath = filepath.Dir(selected.Path)
					}

					// On récupère le shell de l'utilisateur ($SHELL) ou bash par défaut
					shell := os.Getenv("SHELL")
					if shell == "" {
						shell = "/bin/bash"
					}

					// On prépare la commande
					c := exec.Command(shell)
					c.Dir = targetPath

					// tea.ExecProcess suspend l'UI, lance le shell, et reprend l'UI à la fin
					return m, tea.ExecProcess(c, func(err error) tea.Msg {
						// Callback quand le shell est fermé
						return nil
					})
				}
				return m, nil

			// --- NAVIGATION ---
			case "backspace", "left", "h", "esc":
				if m.currentNode.Parent != nil {
					m.currentNode = m.currentNode.Parent
					m.cursor = 0
					m.yOffset = 0
				}

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
	if m.state == StateInputPath {
		title := titleStyle.Render("AED - Analyseur d'Espace Disque")
		input := m.textInput.View()
		return fmt.Sprintf("\n  %s\n\n  Entrez le dossier à analyser :\n  %s\n\n  %s", title, input, helpStyle.Render("(enter: valider • esc: quitter)"))
	}

	if m.state == StateScanning {
		count := atomic.LoadInt64(m.filesScanned)
		return fmt.Sprintf(
			"\n  %s Analyse en cours...\n\n%s fichiers scannés",
			m.spinner.View(),
			countStyle.Render(fmt.Sprintf("%d", count)),
		)
	}

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
		// Ajout de 's' dans le footer
		footer := helpStyle.Render("\n↑/↓/←/→: naviguer • enter: entrer • g: explorer • s: shell • q: quitter")

		return fmt.Sprintf("\n%s\n%s\n%s", header, content, footer)
	}

	return ""
}

// --- LOGIQUE METIER ---

func scanDirectoryCmd(path string, counter *int64, visited map[fileID]struct{}) tea.Cmd {
	return func() tea.Msg {
		root, err := scanRecursively(path, nil, counter, visited)
		return scanFinishedMsg{root: root, err: err}
	}
}

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

	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})

	return node, nil
}

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
