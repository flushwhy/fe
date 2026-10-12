package cmd

import (
	"fmt"

	"github.com/flushwhy/fe/internal/o3de"
	"github.com/flushwhy/fe/internal/ui"
	"github.com/spf13/cobra"
)

var o3deDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose an O3DE engine and basic build prerequisites.",
	RunE: func(cmd *cobra.Command, args []string) error {
		enginePath, _ := cmd.Flags().GetString("engine")
		checks := o3de.Doctor(enginePath)
		failed := 0
		fmt.Println()
		for _, check := range checks {
			if check.OK {
				fmt.Printf("  %s  %-20s %s\n", ui.StyleSuccess.Render(ui.IconOK), ui.StyleBold.Render(check.Name), ui.StyleDim.Render(check.Detail))
				continue
			}
			failed++
			fmt.Printf("  %s  %-20s %s\n", ui.StyleError.Render(ui.IconFail), ui.StyleBold.Render(check.Name), ui.StyleError.Render(check.Detail))
			if check.Fix != "" {
				fmt.Printf("     %s %s\n", ui.StyleDim.Render("next:"), ui.StyleDim.Render(check.Fix))
			}
		}
		fmt.Println()
		if failed > 0 {
			ui.Warn("%d check(s) need attention", failed)
			return fmt.Errorf("O3DE environment is not ready")
		}
		ui.OK("O3DE engine and basic build tools found")
		return nil
	},
}

func init() {
	o3deCmd.AddCommand(o3deDoctorCmd)
	o3deDoctorCmd.Flags().String("engine", "", "path to an O3DE engine root")
}
