package financials

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/antihax/optional"
	mailslurp "github.com/mailslurp/mailslurp-client-go"
)

// create a client and a context
func createClient(apiKey string) (*mailslurp.APIClient, context.Context) {
	// create a context with your api key
	ctx := context.WithValue(
		context.Background(),
		mailslurp.ContextAPIKey,
		mailslurp.APIKey{Key: apiKey},
	)

	// create mailslurp client
	config := mailslurp.NewConfiguration()
	return mailslurp.NewAPIClient(config), ctx
}

// CreateInbox creates a new inbox using the mailslurp client.
// It returns the created inbox and any error that occurred during the API call.
func CreateInbox(apikey string) (mailslurp.InboxDto, error) {
	client, ctx := createClient(apikey)
	opts := &mailslurp.CreateInboxOpts{}
	inbox, response, err := client.InboxControllerApi.CreateInbox(ctx, opts)
	fmt.Println("Create inbox response: ", response)
	return inbox, err
}

func DeleteInbox(apikey string, inboxId string) (int, error) {
	client, ctx := createClient(apikey)
	response, err := client.InboxControllerApi.DeleteInbox(ctx, inboxId)
	fmt.Println("Deleted inbox response: ", response.Body)
	return response.StatusCode, err
}

// getOtpCodeFromMailInbox retrieves the OTP code from a mail inbox.
// It takes a testing.T object and an inboxId string as parameters.
// It creates a client and context, then creates a new inbox using the client.
// It waits for the latest email in the inbox and retrieves its body.
// It uses a regular expression to extract the OTP code from the email body.
// Finally, it prints the OTP code.
// Invite Code Pattern ([0-9]{8})
// Admin OTP Pattern ([0-9]{6})

func GetMailFromMailInbox(inboxId string, pattern string, apikey string) string {
	client, ctx := createClient(apikey)
	inbox, _, err := client.InboxControllerApi.GetInbox(ctx, inboxId)

	if err != nil {
		fmt.Println("Error getting inbox: ", err)
		return ""
	}

	fmt.Println(inbox)
	waitOpts := &mailslurp.WaitForLatestEmailOpts{
		InboxId:    optional.NewInterface(inbox.Id),
		Timeout:    optional.NewInt64(30000),
		UnreadOnly: optional.NewBool(true),
	}
	time.Sleep(60 * time.Second)
	email, response, err := client.WaitForControllerApi.WaitForLatestEmail(ctx, waitOpts)
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}

	fmt.Println("Response: ", response)
	fmt.Println("Email: ", email)
	r := regexp.MustCompile(pattern)
	body := *email.Body
	fmt.Println("Email body: ", body)
	code := r.FindStringSubmatch(body)
	fmt.Println("Code: ", code)
	return code[1]
}

func GetMailFromInbox(inboxId string, pattern string, apikey string, startTime time.Time, timeout bool) (string, error) {
	client, ctx := createClient(apikey)
	inbox, _, _ := client.InboxControllerApi.GetInbox(ctx, inboxId)

	fmt.Println(inbox)
	waitOpts := &mailslurp.WaitForLatestEmailOpts{
		InboxId:    optional.NewInterface(inbox.Id),
		Timeout:    optional.NewInt64(30000),
		UnreadOnly: optional.NewBool(false),
		Since:      optional.NewTime(startTime),
	}
	email, response, err := client.WaitForControllerApi.WaitForLatestEmail(ctx, waitOpts)
	if err != nil {
		fmt.Println("Error: ", err)
		if timeout {
			if strings.Contains(err.Error(), ("408")) {
				return "", nil
			}
		}
		return "", err
	}
	fmt.Println("Response: ", response)
	fmt.Println("Email: ", email)
	r := regexp.MustCompile(pattern)
	body := *email.Subject
	// if startTime.After(email.CreatedAt) {
	// 	return "", err
	// }
	fmt.Println("Email body: ", body)
	code := r.FindStringSubmatch(body)
	fmt.Println("Code: ", code)
	if len(code) <= 0 {
		return "", err
	}
	return code[0], err
}
