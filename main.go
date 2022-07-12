package main

import (
	"bufio"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/option"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	if len(os.Args) != 3 {
		fmt.Printf("Expected [OLD_BILLING_ACCOUNT_ID] [NEW_BILLING_ACCOUNT_ID] as program arguments\n")
		os.Exit(1)
	}

	oldBillingAccountId := os.Args[1]
	newBillingAccountId := os.Args[2]

	projects, err := getProjects(ctx, oldBillingAccountId)
	if err != nil {
		log.Fatal("Error getting projects", err)
	}

	for _, project := range projects {
		fmt.Printf("Moving project [%s] from [%s] to [%s]\n", project.ProjectId, oldBillingAccountId, newBillingAccountId)
		proceed, err := checkProceed()
		if err != nil {
			log.Fatal("Error reading keyboard input", err)
		}
		if proceed {
			err = updateBilling(ctx, project, newBillingAccountId)
			if err != nil {
				log.Fatal("Error updating billing account", err)
			}
			fmt.Printf("Project [%s] moved\n", project.ProjectId)
		} else {
			fmt.Printf("Skipped [%s]\n", project.ProjectId)
		}
	}
}

func checkProceed() (bool, error) {
	fmt.Println("Proceed? (y/n + enter). CTRL+C to quit")
	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	if err != nil {
		return false, err
	}

	switch char {
	case 'y':
		return true, nil
	case 'n':
		return false, nil
	default:
		fmt.Printf("Invalid input [%s]\n", string(char))
		return checkProceed()
	}
}

func updateBilling(ctx context.Context, project *cloudbilling.ProjectBillingInfo, newBillingAccount string) error {
	billingService, err := cloudbilling.NewService(ctx, option.WithScopes(cloudbilling.CloudPlatformScope))
	if err != nil {
		return err
	}

	name := "projects/" + project.ProjectId
	project.BillingAccountName = "billingAccounts/" + newBillingAccount

	_, err = billingService.Projects.UpdateBillingInfo(name, project).Do()
	if err != nil {
		return err
	}

	return nil
}

func getProjects(ctx context.Context, billingAccountId string) ([]*cloudbilling.ProjectBillingInfo, error) {
	var result []*cloudbilling.ProjectBillingInfo

	billingService, err := cloudbilling.NewService(ctx, option.WithScopes(cloudbilling.CloudPlatformScope))
	if err != nil {
		return nil, err
	}

	req := billingService.BillingAccounts.Projects.List("billingAccounts/" + billingAccountId)
	if err := req.Pages(ctx, func(page *cloudbilling.ListProjectBillingInfoResponse) error {
		for _, project := range page.ProjectBillingInfo {
			result = append(result, project)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}
