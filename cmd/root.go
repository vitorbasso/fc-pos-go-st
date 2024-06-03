/*
Copyright © 2024 Vitor Basso
*/
package cmd

import (
	"errors"
	"net/url"
	"os"
	"strest/internal/strest"

	"github.com/spf13/cobra"
)

var (
	ErrMissingURL = errors.New("URL é obrigatória")
)

type RunEFunc func(cmd *cobra.Command, args []string) error

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "strest",
	Short: "Realize um teste de estresse em algum endpoint http",
	Long: `
Permite realizar um teste de estresse em algum endpoint http
com número total de requests a serem realizados e número de
chamadas simultâneas configuráveis.

Ao final, exibe informações sobre o teste como:
 - O tempo total de execução;
 - O número total de requests realizados;
 - A distribuição de status code das respostas;
 - O tempo máximo, mínimo e médio de resposta;
`,
	Example: "strest -u http://localhost:8080 -r 1000 -c 10 -w 5 -H Authorization:'Bearer token',x-correlation-id:123 -m GET",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PreRunE: func(cmd *cobra.Command, args []string) error {
		u, err := cmd.Flags().GetString("url")
		if err != nil {
			return err
		}
		if u == "" {
			return ErrMissingURL
		}
		URL, err := url.Parse(u)
		if err != nil {
			return err
		}
		cmd.Flags().Set("url", URL.String())
		r, err := cmd.Flags().GetInt("requests")
		if err != nil {
			return err
		}
		if r < 1 {
			err = cmd.Flags().Set("requests", "1")
			if err != nil {
				return err
			}
		}
		c, err := cmd.Flags().GetInt("concurrency")
		if err != nil {
			return err
		}
		if c < 1 {
			err = cmd.Flags().Set("concurrency", "1")
			if err != nil {
				return err
			}
		}
		t, err := cmd.Flags().GetInt("timeout")
		if err != nil {
			return err
		}
		if t < 0 {
			err = cmd.Flags().Set("timeout", "0")
			if err != nil {
				return err
			}
		}
		w, err := cmd.Flags().GetInt("warmup")
		if err != nil {
			return err
		}
		if w < 0 {
			err = cmd.Flags().Set("warmup", "0")
			if err != nil {
				return err
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		headers, _ := cmd.Flags().GetStringArray("header")
		timeout, _ := cmd.Flags().GetInt("timeout")
		requests, _ := cmd.Flags().GetInt("requests")
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		warmup, _ := cmd.Flags().GetInt("warmup")

		strestConfig := &strest.Configuration{
			URL:         cmd.Flag("url").Value.String(),
			Method:      cmd.Flag("method").Value.String(),
			Headers:     headers,
			Body:        cmd.Flag("body").Value.String(),
			Timeout:     timeout,
			Requests:    requests,
			Concurrency: concurrency,
			Warmup:      warmup,
		}
		strest := strest.NewStrest(strestConfig)
		result, err := strest.Run()
		if err != nil {
			return err
		}
		cmd.Println(result.String())
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("url", "u", "", "URL do serviço a ser testado.")
	rootCmd.MarkFlagRequired("url")
	rootCmd.Flags().IntP("requests", "r", 1, "Número total de requests.")
	rootCmd.Flags().IntP("concurrency", "c", 1, "Número de chamadas simultâneas.")
	rootCmd.Flags().StringP("method", "m", "GET", "Método HTTP a ser utilizado.")
	rootCmd.Flags().StringP("body", "b", "", "Corpo da requisição.")
	rootCmd.Flags().StringArrayP("header", "H", []string{}, "Header da requisição.")
	rootCmd.Flags().IntP("timeout", "t", 10, "Timeout da requisição em segundos.")
	rootCmd.Flags().IntP("warmup", "w", 0, "Número de requisições a serem realizadas antes do teste em si.")
}
