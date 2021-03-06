package commands

import (
	"fmt"

	"github.com/argoproj/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj/argo/cmd/argo/commands/client"
	workflowpkg "github.com/argoproj/argo/pkg/apiclient/workflow"
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
)

// NewDeleteCommand returns a new instance of an `argo delete` command
func NewDeleteCommand() *cobra.Command {
	var (
		flags         listFlags
		all           bool
		allNamespaces bool
		dryRun        bool
	)
	var command = &cobra.Command{
		Use:   "delete [--dry-run] [WORKFLOW...|[--all] [--older] [--completed] [--prefix PREFIX] [--selector SELECTOR]]",
		Short: "delete workflows",
		Example: `# Delete a workflow:

  argo delete my-wf

# Delete the latest workflow:

  argo delete @latest
`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx, apiClient := client.NewAPIClient()
			serviceClient := apiClient.NewWorkflowServiceClient()
			var workflows wfv1.Workflows
			if !allNamespaces {
				flags.namespace = client.Namespace()
			}
			for _, name := range args {
				workflows = append(workflows, wfv1.Workflow{
					ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: flags.namespace},
				})
			}
			if all || flags.completed || flags.prefix != "" || flags.labels != "" {
				listed, err := listWorkflows(ctx, serviceClient, flags)
				errors.CheckError(err)
				workflows = append(workflows, listed...)
			}
			for _, wf := range workflows {
				if !dryRun {
					_, err := serviceClient.DeleteWorkflow(ctx, &workflowpkg.WorkflowDeleteRequest{Name: wf.Name, Namespace: wf.Namespace})
					if err != nil && status.Code(err) == codes.NotFound {
						fmt.Printf("Workflow '%s' not found\n", wf.Name)
						continue
					}
					errors.CheckError(err)
					fmt.Printf("Workflow '%s' deleted\n", wf.Name)
				} else {
					fmt.Printf("Workflow '%s' deleted (dry-run)\n", wf.Name)
				}
			}
		},
	}

	command.Flags().BoolVar(&allNamespaces, "all-namespaces", false, "Delete workflows from all namespaces")
	command.Flags().BoolVar(&all, "all", false, "Delete all workflows")
	command.Flags().BoolVar(&flags.completed, "completed", false, "Delete completed workflows")
	command.Flags().StringVar(&flags.prefix, "prefix", "", "Delete workflows by prefix")
	command.Flags().StringVar(&flags.finishedAfter, "older", "", "Delete completed workflows finished before the specified duration (e.g. 10m, 3h, 1d)")
	command.Flags().StringVarP(&flags.labels, "selector", "l", "", "Selector (label query) to filter on, not including uninitialized ones")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Do not delete the workflow, only print what would happen")
	return command
}
