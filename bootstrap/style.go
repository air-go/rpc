package bootstrap

import (
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	width = 80
)

// Style definitions.
var (
	subtle  = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	// Title.
	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			SetString("AirGo")

	url = lipgloss.NewStyle().Foreground(special).Render

	descStyle = lipgloss.NewStyle().MarginTop(1)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle)

	// .
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	pidStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF1493"))

	docStyle = lipgloss.NewStyle().Padding(1, 1, 0, 1)
)

func welcome() string {
	doc := strings.Builder{}

	// title
	{
		var (
			colors = colorGrid(1, 5)
			title  strings.Builder
		)

		for i, v := range colors {
			const offset = 2
			c := lipgloss.Color(v[0])
			fmt.Fprint(&title, titleStyle.Copy().MarginLeft(i*offset).Background(c))
			if i < len(colors)-1 {
				title.WriteRune('\n')
			}
		}

		desc := lipgloss.JoinVertical(lipgloss.Center,
			descStyle.Render("Out-Of-The-Box"),
			infoStyle.Render("Source Code"+divider+url("https://github.com/air-go/rpc")),
		)

		row := lipgloss.JoinHorizontal(lipgloss.Top, title.String(), desc)
		doc.WriteString(row + "\n")
	}

	// hello
	{
		color := lipgloss.AdaptiveColor{Light: "#8B8989", Dark: "#FFFAFA"}
		hello := lipgloss.NewStyle().Foreground(color).Bold(true).Width(60).Align(lipgloss.Center).Render("Hello AirGo")
		ui := lipgloss.JoinVertical(lipgloss.Center, hello)

		dialog := lipgloss.Place(width, 0,
			lipgloss.Center, lipgloss.Center,
			dialogBoxStyle.Render(ui),
			lipgloss.WithWhitespaceForeground(subtle),
		)

		doc.WriteString(dialog + "\n")
	}

	{
		pid := fmt.Sprintf("%s Actual pid is %d\n", time.Now().Format("2006-01-02 15:04:05"), syscall.Getpid())
		doc.WriteString(pidStyle.Render(pid))
	}

	return docStyle.Render(doc.String())
}

func colorGrid(xSteps, ySteps int) [][]string {
	x0y0, _ := colorful.Hex("#87CEFA")
	x1y0, _ := colorful.Hex("#00BFFF")
	x0y1, _ := colorful.Hex("#1E90FF")
	x1y1, _ := colorful.Hex("#0000FF")

	x0 := make([]colorful.Color, ySteps)
	for i := range x0 {
		x0[i] = x0y0.BlendLuv(x0y1, float64(i)/float64(ySteps))
	}

	x1 := make([]colorful.Color, ySteps)
	for i := range x1 {
		x1[i] = x1y0.BlendLuv(x1y1, float64(i)/float64(ySteps))
	}

	grid := make([][]string, ySteps)
	for x := 0; x < ySteps; x++ {
		y0 := x0[x]
		grid[x] = make([]string, xSteps)
		for y := 0; y < xSteps; y++ {
			grid[x][y] = y0.BlendLuv(x1[x], float64(y)/float64(xSteps)).Hex()
		}
	}

	return grid
}
