package schemaherokubectlcli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/schemahero/schemahero/pkg/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply a spec to a database",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			// to support automaticenv, we can't use cobra required flags
			driver := v.GetString("driver")
			ddl := v.GetString("ddl")
			uri := v.GetString("uri")

			if driver == "" || ddl == "" || uri == "" {
				missing := []string{}
				if driver == "" {
					missing = append(missing, "driver")
				}
				if ddl == "" {
					missing = append(missing, "ddl")
				}
				if uri == "" {
					missing = append(missing, "uri")
				}

				return fmt.Errorf("missing required params: %v", missing)
			}

			fi, err := os.Stat(v.GetString("ddl"))
			if err != nil {
				return err
			}

			db := database.Database{
				InputDir:  v.GetString("input-dir"),
				OutputDir: v.GetString("output-dir"),
				Driver:    v.GetString("driver"),
				URI:       v.GetString("uri"),
			}

			if fi.Mode().IsDir() {
				commands := []string{}
				err := filepath.Walk(v.GetString("ddl"), func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}

					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()

					commands := []string{}
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						commands = append(commands, scanner.Text())
					}

					if err := scanner.Err(); err != nil {
						return err
					}

					return nil
				})

				if err != nil {
					return err
				}

				if err := db.ApplySync(commands); err != nil {
					return err
				}

				return nil
			} else {
				f, err := os.Open(v.GetString("ddl"))
				if err != nil {
					return err
				}
				defer f.Close()

				commands := []string{}
				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					commands = append(commands, scanner.Text())
				}

				if err := scanner.Err(); err != nil {
					return err
				}

				return db.ApplySync(commands)
			}
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("uri", "", "connection string uri to use")
	cmd.Flags().String("ddl", "", "filename or directory name containing the rendered DDL commands to execute")

	return cmd
}
