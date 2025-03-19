package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type model struct {
	capacity int
	charging bool
	icon     string
}

func findBatteryPath() string {
	batteries, _ := filepath.Glob("/sys/class/power_supply/BAT*")
	for _, bat := range batteries {
		if _, err := os.Stat(filepath.Join(bat, "uevent")); err == nil {
			return bat
		}
	}
	return ""
}

func readBatteryInfo(batteryPath string) (int, bool, map[string]string) {
	data, err := os.ReadFile(filepath.Join(batteryPath, "uevent"))
	if err != nil {
		fmt.Println("Error reading battery info:", err)
		os.Exit(1)
	}

	info := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			info[parts[0]] = parts[1]
		}
	}
	capacity, _ := strconv.Atoi(info["POWER_SUPPLY_CAPACITY"])
	charging := info["POWER_SUPPLY_STATUS"] == "Charging" || info["POWER_SUPPLY_STATUS"] == "Full"

	return capacity, charging, info
}

func getBatteryIcon(level int, charging bool) string {
	iconsCharging := map[int]string{
		100: "🔋⚡", 90: "🔋⚡", 80: "🔋⚡", 70: "🔋⚡", 60: "🔋⚡", 50: "🔋⚡",
		40: "🔋⚡", 30: "🔋⚡", 20: "🔋⚡", 10: "🔋⚡", 0: "🔋⚡"}
	icons := map[int]string{
		100: "🔋", 90: "🔋", 80: "🔋", 70: "🔋", 60: "🔋", 50: "🔋",
		40: "🔋", 30: "🔋", 20: "🪫", 10: "🪫", 0: "⚠️"}

	//🔌
	var selectedIcons map[int]string
	if charging {
		selectedIcons = iconsCharging
	} else {
		selectedIcons = icons
	}

	thresholds := []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10, 0}
	for _, t := range thresholds {
		if level >= t {
			return selectedIcons[t]
		}
	}
	return "❓"
}
func getBatteryIconTui(level int, charging bool) string {
	var icon string
	if charging {
		icon = "Charging: "
	} else {
		icon = "Battery: "
	}

	// Create a simple battery icon using ASCII characters
	fullBlocks := level / 10
	emptyBlocks := 10 - fullBlocks
	icon += "["
	for i := 0; i < fullBlocks; i++ {
		icon += "="
	}
	for i := 0; i < emptyBlocks; i++ {
		icon += " "
	}
	icon += "]"

	return icon
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return struct{}{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case struct{}:
		batteryPath := findBatteryPath()
		if batteryPath == "" {
			return m, tea.Quit
		}
		m.capacity, m.charging, _ = readBatteryInfo(batteryPath)
		m.icon = getBatteryIcon(m.capacity, m.charging)
		return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
			return struct{}{}
		})
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Battery: %s %d%%\n", m.icon, m.capacity)
}

func main() {
	var showIcon, showTuiIcon, showPercentage, number bool

	var rootCmd = &cobra.Command{Use: "batstat"}

	var statsCmd = &cobra.Command{
		Use:     "stats [s]",
		Aliases: []string{"s"},
		Short:   "Display battery stats",
		Run: func(cmd *cobra.Command, args []string) {
			batteryPath := findBatteryPath()
			if batteryPath == "" {
				fmt.Println("No battery found.")
				os.Exit(1)
			}
			capacity, charging, _ := readBatteryInfo(batteryPath)
			var icon string
			if showIcon {
				icon = getBatteryIcon(capacity, charging)
			} else if showTuiIcon {
				icon = getBatteryIconTui(capacity, charging)
			}
			var output strings.Builder

			if number {
				if showIcon || showTuiIcon {
					output.WriteString(fmt.Sprintf("%s ", icon))
				}

			} else {
				for _, arg := range args {
					if arg == "-i" {
						output.WriteString(fmt.Sprintf("%s ", icon))
					}
					if arg == "-p" {
						output.WriteString(fmt.Sprintf("%d%% ", capacity))
					}
				}
				if showIcon || showTuiIcon {
					output.WriteString(fmt.Sprintf("%s ", icon))
				}
				if showPercentage {
					output.WriteString(fmt.Sprintf("%d%%", capacity))
				}
				if (!showIcon || !showTuiIcon) && !showPercentage {
					output.WriteString(fmt.Sprintf("%d", capacity))
				}
			}

			fmt.Println(output.String())
		},
	}

	var infoCmd = &cobra.Command{
		Use:     "info [i]",
		Aliases: []string{"i"},
		Short:   "Display detailed battery information",
		Run: func(cmd *cobra.Command, args []string) {
			batteryPath := findBatteryPath()
			if batteryPath == "" {
				fmt.Println("No battery found.")
				os.Exit(1)
			}
			capacity, charging, info := readBatteryInfo(batteryPath)
			icon := getBatteryIcon(capacity, charging)
			fmt.Printf("Battery: %s %d%%\n", icon, capacity)
			for key, value := range info {
				fmt.Printf("%s: %s\n", key, value)
			}
		},
	}

	statsCmd.Flags().BoolVarP(&showIcon, "icon", "i", false, "Show battery icon")
	statsCmd.Flags().BoolVarP(&showTuiIcon, "tui icon", "t", false, "Show tui battery icon")
	statsCmd.Flags().BoolVarP(&showPercentage, "percentage", "p", false, "Show battery percentage")
	statsCmd.Flags().BoolVarP(&number, "number", "n", false, "Show battery number only")

	rootCmd.AddCommand(statsCmd, infoCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	}
}

// todo separar
// todo crear um watcher/dam
