package cmd

import (
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cobra"
)

func findCommand(commands []*cobra.Command, name string) *cobra.Command {
	for _, command := range commands {
		if command.Name() == name {
			return command
		}
	}
	return nil
}

func TestLoad(t *testing.T) {
	Convey("Given the loaded root command", t, func() {
		root, err := Load()
		So(err, ShouldBeNil)

		compareCmd := findCommand(root.Commands(), "compare")
		exportCmd := findCommand(root.Commands(), exportCommandName)
		importCmd := findCommand(root.Commands(), importCommandName)

		Convey("Then compare, export and import are registered as sibling top-level commands", func() {
			So(compareCmd, ShouldNotBeNil)
			So(exportCmd, ShouldNotBeNil)
			So(importCmd, ShouldNotBeNil)
		})

		Convey("Then export is not nested as a subcommand of compare", func() {
			So(compareCmd.Commands(), ShouldBeEmpty)
		})

		Convey("Then export requires the output flag and import requires the input flag", func() {
			So(exportCmd.Flag("output"), ShouldNotBeNil)
			So(importCmd.Flag("input"), ShouldNotBeNil)
		})

		Convey("Then the verbose persistent flag is defined", func() {
			So(root.PersistentFlags().Lookup("verbose"), ShouldNotBeNil)
		})
	})
}

func TestExportCommandRequiresOutput(t *testing.T) {
	Convey("Given an export command without an output path", t, func() {
		command, err := exportCommand()
		So(err, ShouldBeNil)
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)

		Convey("When the command is executed", func() {
			err := command.Execute()

			Convey("Then it should reject the missing required flag", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, `required flag(s) "output" not set`)
			})
		})
	})
}

func TestImportCommandRequiresInput(t *testing.T) {
	Convey("Given an import command without an input path", t, func() {
		command, err := importCommand()
		So(err, ShouldBeNil)
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)

		Convey("When the command is executed", func() {
			err := command.Execute()

			Convey("Then it should reject the missing required flag", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, `required flag(s) "input" not set`)
			})
		})
	})
}
