package sqltui

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
)

// Message de retour au menu principal
type BackMsg struct{}

type SessionState int

const (
	StateSelectFile SessionState = iota
	StateBrowser
)

type FocusArea int

const (
	FocusList FocusArea = iota
	FocusData
)

type tableItem struct {
	name, desc string
}

func (i tableItem) Title() string       { return i.name }
func (i tableItem) Description() string { return i.desc }
func (i tableItem) FilterValue() string { return i.name }

// --- Styles ---
var (
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true)
	borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#500aff"))
	focusStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#00f6ff"))

	// AJOUTS POUR L'HARMONISATION UX
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00d4"))            // Gris pour l'aide
	pathStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#500aff")).Bold(true) // Cyan pour le chemin
)

type Model struct {
	state         SessionState
	focus         FocusArea
	filePicker    filepicker.Model
	list          list.Model
	table         table.Model
	db            *sql.DB
	dbPath        string
	currentTable  string
	width, height int
	err           error
}

func New(w, h int) Model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".db", ".sqlite", ".sqlite3", ".db3"}
	fp.CurrentDirectory, _ = os.Getwd()

	// --- PERSONNALISATION DES COULEURS (Comme les autres outils) ---
	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D"))              // Rose
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF2A6D")).Bold(true) // Rose Gras
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(lipgloss.Color("#00f6ff"))           // Cyan
	fp.Styles.File = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))                // Blanc
	// -------------------------------------------------------------

	// Ajustement hauteur pour header/footer
	fpHeight := h - 8
	if fpHeight < 5 {
		fpHeight = 5
	}
	fp.Height = fpHeight
	fp.ShowHidden = false

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 30, h-5)
	l.Title = "Tables"
	l.SetShowHelp(false)

	t := table.New(
		table.WithColumns([]table.Column{{Title: "Info", Width: 20}}),
		table.WithRows([]table.Row{{"Sélectionnez une table..."}}),
		table.WithFocused(false),
		table.WithHeight(h-5),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(true)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("#FF2A6D")).Bold(false)
	t.SetStyles(s)

	return Model{
		state:      StateSelectFile,
		focus:      FocusList,
		filePicker: fp,
		list:       l,
		table:      t,
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
		m.filePicker.Height = msg.Height - 8 // Ajustement hauteur
		m.list.SetHeight(msg.Height - 5)
		m.table.SetHeight(msg.Height - 8)

	case tea.KeyMsg:
		if m.state == StateSelectFile {
			switch msg.String() {
			case "q":
				return m, func() tea.Msg { return BackMsg{} }
			case "h": // Toggle fichiers cachés
				m.filePicker.ShowHidden = !m.filePicker.ShowHidden
				return m, m.filePicker.Init()
			}
		}

		if m.state == StateBrowser && msg.String() == "ctrl+c" {
			if m.db != nil {
				m.db.Close()
			}
			return m, func() tea.Msg { return BackMsg{} }
		}
	}

	switch m.state {
	case StateSelectFile:
		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)
		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			m.dbPath = path
			err := m.openDB(path)
			if err != nil {
				m.err = err
			} else {
				m.state = StateBrowser
				m.loadTables()
			}
		}

	case StateBrowser:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "tab":
				if m.focus == FocusList {
					m.focus = FocusData
					m.table.Focus()
				} else {
					m.focus = FocusList
					m.table.Blur()
				}
			case "enter":
				if m.focus == FocusList {
					selectedItem := m.list.SelectedItem()
					if selectedItem != nil {
						tableName := selectedItem.(tableItem).name
						m.currentTable = tableName
						m.loadTableData(tableName)
					}
				}
			case "esc":
				if m.focus == FocusList && !m.list.SettingFilter() {
					if m.db != nil {
						m.db.Close()
					}
					m.state = StateSelectFile
					return m, m.filePicker.Init()
				}
			}
		}

		if m.focus == FocusList {
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	// --- VUE 1 : SELECTEUR DE FICHIER (mise à jour UX) ---
	if m.state == StateSelectFile {
		// 1. Titre
		title := titleStyle.Render("SqlTUI - Ouvrir une base de données")

		// 2. Chemin actuel (PWD)
		currentDir := fmt.Sprintf(" %s", pathStyle.Render(m.filePicker.CurrentDirectory))

		// 3. Contenu Picker
		content := "\n" + m.filePicker.View()

		// 4. Pied de page : Aide
		hiddenStatus := "off"
		if m.filePicker.ShowHidden {
			hiddenStatus = "ON"
		}
		helpText := fmt.Sprintf("↑/↓: naviguer • enter: ouvrir • h: cachés(%s) • q: retour", hiddenStatus)
		footer := helpStyle.Render("\n" + helpText)

		return fmt.Sprintf("\n  %s\n\n  %s%s%s", title, currentDir, content, footer)
	}

	// --- VUE 2 : NAVIGATEUR BDD ---
	if m.state == StateBrowser {
		header := titleStyle.Render(fmt.Sprintf(" BDD: %s ", m.dbPath))

		listStyle := borderStyle
		if m.focus == FocusList {
			listStyle = focusStyle
		}
		listView := listStyle.Width(30).Height(m.height - 4).Render(m.list.View())

		tableStyle := borderStyle
		if m.focus == FocusData {
			tableStyle = focusStyle
		}

		// Calcul de sécurité
		safeTableWidth := m.width - 36
		if safeTableWidth < 10 {
			safeTableWidth = 10
		}

		tableInfo := fmt.Sprintf("Table: %s", m.currentTable)
		if m.currentTable == "" {
			tableInfo = "Aucune table sélectionnée"
		}

		tableView := tableStyle.
			Width(safeTableWidth).
			MaxWidth(safeTableWidth). // Sécurité max width
			Height(m.height - 4).
			Render(lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Bold(true).Render(tableInfo),
				m.table.View(),
			))

		body := lipgloss.JoinHorizontal(lipgloss.Top, listView, tableView)

		helpText := "tab: changer de vue • enter: charger table • /: chercher table • q/esc: retour"
		if m.focus == FocusData {
			helpText = "tab: retour liste • ↑/↓: naviguer données"
		}
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(helpText)

		return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	}

	return "Loading..."
}

func (m *Model) openDB(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *Model) loadTables() {
	if m.db == nil {
		return
	}
	rows, err := m.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return
	}
	defer rows.Close()

	var items []list.Item
	for rows.Next() {
		var name string
		rows.Scan(&name)
		items = append(items, tableItem{name: name, desc: "Table SQLite"})
	}
	m.list.SetItems(items)
}

func (m *Model) loadTableData(tableName string) {
	if m.db == nil {
		return
	}

	// 1. Reset ROWS
	m.table.SetRows([]table.Row{})

	// 2. Récupérer les colonnes d'abord
	dummyQuery := fmt.Sprintf("SELECT * FROM %s LIMIT 1", tableName)
	rows, err := m.db.Query(dummyQuery)
	if err != nil {
		return
	}

	allColumns, err := rows.Columns()
	rows.Close()
	if err != nil {
		return
	}

	// 3. LOGIQUE DE SECURITE D'AFFICHAGE
	availableWidth := m.width - 40
	if availableWidth < 0 {
		availableWidth = 10
	}

	const minColWidth = 12
	maxColsToDisplay := availableWidth / minColWidth
	if maxColsToDisplay < 1 {
		maxColsToDisplay = 1
	}

	// Si la table a trop de colonnes, on coupe
	displayColumns := allColumns
	if len(allColumns) > maxColsToDisplay {
		displayColumns = allColumns[:maxColsToDisplay]
	}

	// Configuration des colonnes
	var tableCols []table.Column
	finalColWidth := availableWidth / len(displayColumns)

	for _, col := range displayColumns {
		tableCols = append(tableCols, table.Column{Title: col, Width: finalColWidth})
	}
	m.table.SetColumns(tableCols)

	// 4. Récupérer les vraies données
	safeCols := strings.Join(displayColumns, ", ")
	finalQuery := fmt.Sprintf("SELECT %s FROM %s LIMIT 50", safeCols, tableName)

	dataRows, err := m.db.Query(finalQuery)
	if err != nil {
		return
	}
	defer dataRows.Close()

	var tableRows []table.Row
	values := make([]interface{}, len(displayColumns))
	valuePtrs := make([]interface{}, len(displayColumns))

	for dataRows.Next() {
		for i := range displayColumns {
			valuePtrs[i] = &values[i]
		}
		dataRows.Scan(valuePtrs...)

		var rowData []string
		for _, val := range values {
			var v string
			if val == nil {
				v = "NULL"
			} else {
				switch t := val.(type) {
				case []byte:
					v = string(t)
				default:
					v = fmt.Sprintf("%v", t)
				}
			}
			rowData = append(rowData, v)
		}
		tableRows = append(tableRows, rowData)
	}

	m.table.SetRows(tableRows)
	m.table.GotoTop()
}
