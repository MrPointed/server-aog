package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateCharts(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) viewCharts() string {

	if m.chartsData.Err != nil {

		return fmt.Sprintf("Error fetching charts data: %v", m.chartsData.Err)

	}



	// 1. Users per Hour (Vertical Bar Chart)

	historyH := m.chartsData.HistoryHourly

	if len(historyH) < 24 {

		newH := make([]int, 24)

		copy(newH, historyH)

		historyH = newH

	}



	maxH := 0

	for _, v := range historyH {

		if v > maxH { maxH = v }

	}

	if maxH == 0 { maxH = 1 }



	chartHeight := 10

	var hRows []string

	hRows = append(hRows, titleStyle.Render("[Users per Hour (last 24h)]"))



	for r := chartHeight; r > 0; r-- {

		// Y-axis label

		label := ""

		if r == chartHeight {

			label = fmt.Sprintf("%3d ┐", maxH)

		} else if r == 1 {

			label = "  0 ┘"

		} else {

			label = "    │"

		}

		

				row := label + " "

		

				for h := 0; h < 24; h++ {

		

					val := historyH[h]

		

					barHeight := (val * chartHeight) / maxH

		

					// Ensure some visibility for values > 0

		

					if val > 0 && barHeight == 0 { barHeight = 1 }

		

		

		

					if barHeight >= r {

		

						row += lipgloss.NewStyle().Foreground(special).Render(" █")

		

					} else {

		

						row += "  "

		

					}

		

					// Add extra space to reach ~60 chars (24*2.5 = 60)

		

					if h%2 == 1 {

		

						row += " "

		

					}

		

				}

		

				hRows = append(hRows, row)

		

			}

		

			

		

			// X-axis (matching the ~60 char width)

		

			xAxis := "      " + strings.Repeat("▔", 60)

		

			hRows = append(hRows, xAxis)

		

			

		

			xLabels := "      "

		

			for h := 0; h < 24; h++ {

		

				if h%2 == 0 {

		

					xLabels += fmt.Sprintf("%02d", h)

		

				} else {

		

					// spacing to keep aligned with bars

		

					xLabels += " "

		

				}

		

				// extra space every other hour to match bars

		

				if h%2 == 1 {

		

					xLabels += "  "

		

				}

		

			}

		

			hRows = append(hRows, xLabels)



		barChartH := strings.Join(hRows, "\n")



	



		// 2. Class Distribution (Vertical Bar Chart)



		dist := m.chartsData.ClassDistribution



		barChartC := ""



		if len(dist) > 0 {



			type classCount struct {



				Class string



				Count int



			}



			var sorted []classCount



			for k, v := range dist {



				sorted = append(sorted, classCount{k, v})



			}



			sort.Slice(sorted, func(i, j int) bool {



				return sorted[i].Count > sorted[j].Count



			})



	



			maxC := 0



			for _, item := range sorted {



				if item.Count > maxC { maxC = item.Count }



			}



			if maxC == 0 { maxC = 1 }



	



			var cRows []string



			cRows = append(cRows, "\n"+titleStyle.Render("[All-time Class Distribution (Lvl 25+)]"))



			



			numClasses := len(sorted)



			for r := chartHeight; r > 0; r-- {



				label := ""



				if r == chartHeight {



					label = fmt.Sprintf("%3d ┐", maxC)



				} else if r == 1 {



					label = "  0 ┘"



				} else {



					label = "    │"



				}



				



				row := label + " "



				for i := 0; i < numClasses; i++ {



					val := sorted[i].Count



					barHeight := (val * chartHeight) / maxC



					if val > 0 && barHeight == 0 { barHeight = 1 }



	



					if barHeight >= r {



						row += lipgloss.NewStyle().Foreground(special).Render("  █  ")



					} else {



						row += "     "



					}



				}



				cRows = append(cRows, row)



			}



			cRows = append(cRows, "      "+strings.Repeat("▔", numClasses*5))



			



			cLabels := "      "



			for i := 0; i < numClasses; i++ {



				name := sorted[i].Class



				if len(name) > 4 {



					name = name[:4]



				}



				cLabels += fmt.Sprintf(" %-4s", name)



			}



			cRows = append(cRows, cLabels)



			barChartC = strings.Join(cRows, "\n")



		}



	



		// 3. Users per Day (Vertical Bar Chart)



		historyD := m.chartsData.HistoryDaily

	barChartD := ""

	if len(historyD) > 0 {

		maxD := 0

		for _, v := range historyD {

			if v > maxD { maxD = v }

		}

		if maxD == 0 { maxD = 1 }



		var dRows []string

		dRows = append(dRows, "\n"+titleStyle.Render("[Users per Day (last 30d)]"))

		

		dWidth := len(historyD)

		for r := chartHeight; r > 0; r-- {

			label := ""

			if r == chartHeight {

				label = fmt.Sprintf("%3d ┐", maxD)

			} else if r == 1 {

				label = "  0 ┘"

			} else {

				label = "    │"

			}

			

			row := label + " "

			for d := 0; d < dWidth; d++ {

				val := historyD[d]

				barHeight := (val * chartHeight) / maxD

				if val > 0 && barHeight == 0 { barHeight = 1 }



				if barHeight >= r {

					row += lipgloss.NewStyle().Foreground(special).Render("█")

				} else {

					row += " "

				}

			}

			dRows = append(dRows, row)

		}

		dRows = append(dRows, "      "+strings.Repeat("▔", dWidth))

		

		dLabels := "      "

		for d := 0; d < dWidth; d++ {

			if d%5 == 0 {

				dLabels += fmt.Sprintf("D%02d", d)

				d += 2

			} else {

				dLabels += " "

			}

		}

				dRows = append(dRows, dLabels)

				barChartD = strings.Join(dRows, "\n")

			}

		

			return detailStyle.Render(barChartH + barChartC + barChartD)

		}
