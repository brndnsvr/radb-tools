package cli

import (
	"context"
	"fmt"

	"github.com/bss/radb-client/internal/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewSearchCmd creates the search command and its subcommands.
func NewSearchCmd(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Aliases: []string{"find"},
		Short:   "Search for objects in RADb",
		Long:    "Search for routes, contacts, AS-sets, and other objects",
	}

	cmd.AddCommand(
		newSearchQueryCmd(logger),
		newSearchValidateASNCmd(logger),
	)

	return cmd
}

// newSearchQueryCmd creates the search query command.
func newSearchQueryCmd(logger *logrus.Logger) *cobra.Command {
	var (
		outputFormat string
		objectType   string
	)

	cmd := &cobra.Command{
		Use:   "query <search-term>",
		Short: "Search for objects",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()
			query := args[0]

			// Use the shared API client from CLI context (already authenticated)
			results, err := ctx.APIClient.Search(cmdCtx, query, objectType)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			switch outputFormat {
			case "json":
				return outputter.renderJSON(results)
			case "yaml":
				return outputter.renderYAML(results)
			default:
				// Handle both JSON (SearchResult) and RPSL (map) responses
				if searchResult, ok := results.(*api.SearchResult); ok {
					// JSON format response
					fmt.Printf("Found %d results for query: %s\n\n", searchResult.Count, searchResult.Query)
					for i, result := range searchResult.Results {
						fmt.Printf("%d. ", i+1)
						for key, value := range result {
							fmt.Printf("%s=%v ", key, value)
						}
						fmt.Println()
					}
				} else if rawMap, ok := results.(map[string]interface{}); ok {
					// RPSL format response
					if rawResponse, ok := rawMap["raw_response"].(string); ok {
						fmt.Println(rawResponse)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVarP(&objectType, "type", "t", "", "Object type (route, contact, as-set, etc.)")

	return cmd
}

// newSearchValidateASNCmd creates the validate asn command.
func newSearchValidateASNCmd(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-asn <asn>",
		Short: "Validate an ASN",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()
			asn := args[0]

			// Use the shared API client from CLI context (already authenticated)
			valid, err := ctx.APIClient.ValidateASN(cmdCtx, asn)
			if err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			if valid {
				fmt.Printf("ASN %s is valid\n", asn)
			} else {
				fmt.Printf("ASN %s is NOT valid\n", asn)
			}

			return nil
		},
	}

	return cmd
}
