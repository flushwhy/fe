/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// resizetextureCmd represents the resizetexture command
var resizetextureCmd = &cobra.Command{
	Use:   "resizetexture",
	Short: "Take your textures from 16k to 2k.",
	Long:  `This command is used to resize textures from everything from 16k to 2k.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("resizetexture called")
	},
}

func init() {
	rootCmd.AddCommand(resizetextureCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resizetextureCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resizetextureCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
