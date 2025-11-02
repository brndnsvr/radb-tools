package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// NewCsqrCmd creates the csqr command for CenterSquare-specific operations.
func NewCsqrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "csqr-all",
		Short: "List all routes for CenterSquare maintainers (MAINT-AS32298 and MAINT-AS12213)",
		Long: `Query all route objects maintained by CenterSquare's maintainer objects:
  - MAINT-AS32298 (Evoque Data Center Solutions)
  - MAINT-AS12213 (Cyxtera)

This is equivalent to running:
  radb-client search query -- "-i mnt-by MAINT-AS32298"
  radb-client search query -- "-i mnt-by MAINT-AS12213"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()

			// Query for MAINT-AS32298
			fmt.Println("# Routes maintained by MAINT-AS32298 (Evoque Data Center Solutions)")
			fmt.Println("# ======================================================================")
			fmt.Println()

			results1, err := ctx.APIClient.Search(cmdCtx, "-i mnt-by MAINT-AS32298", "")
			if err != nil {
				return fmt.Errorf("failed to query MAINT-AS32298: %w", err)
			}

			// Display results for MAINT-AS32298
			if rawMap, ok := results1.(map[string]interface{}); ok {
				if rawResponse, ok := rawMap["raw_response"].(string); ok {
					fmt.Println(rawResponse)
				}
			}

			fmt.Println()
			fmt.Println()
			fmt.Println("# Routes maintained by MAINT-AS12213 (Cyxtera)")
			fmt.Println("# ======================================================================")
			fmt.Println()

			// Query for MAINT-AS12213
			results2, err := ctx.APIClient.Search(cmdCtx, "-i mnt-by MAINT-AS12213", "")
			if err != nil {
				return fmt.Errorf("failed to query MAINT-AS12213: %w", err)
			}

			// Display results for MAINT-AS12213
			if rawMap, ok := results2.(map[string]interface{}); ok {
				if rawResponse, ok := rawMap["raw_response"].(string); ok {
					fmt.Println(rawResponse)
				}
			}

			return nil
		},
	}

	return cmd
}
