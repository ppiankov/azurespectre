package commands

import (
	"fmt"
	"strings"
)

func enhanceError(action string, err error) error {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "DefaultAzureCredential"):
		return fmt.Errorf("%s: %w\n\nHint: Configure Azure credentials. Run 'az login' or set AZURE_CLIENT_ID/AZURE_TENANT_ID/AZURE_CLIENT_SECRET", action, err)
	case strings.Contains(msg, "AuthorizationFailed") || strings.Contains(msg, "AuthenticationFailed"):
		return fmt.Errorf("%s: %w\n\nHint: Ensure your account has Reader role on the subscription. Run 'azurespectre init' to generate the required role definition", action, err)
	case strings.Contains(msg, "SubscriptionNotFound"):
		return fmt.Errorf("%s: %w\n\nHint: Check that the subscription ID is correct and you have access to it", action, err)
	case strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "context canceled"):
		return fmt.Errorf("%s: %w\n\nHint: Scan timed out. Try --timeout with a longer duration or limit scope with --resource-group", action, err)
	default:
		return fmt.Errorf("%s: %w", action, err)
	}
}
