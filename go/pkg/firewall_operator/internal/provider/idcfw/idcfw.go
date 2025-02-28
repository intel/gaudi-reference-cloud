// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Package idcfw provides primitives to interact with the openapi HTTP API.
package idcfw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/internal/provider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

// UserSchema defines model for UserSchema.
type UserSchema struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Defines values for EnvironmentZone.
const (
	Environment_Prod  provider.EnvironmentZone = "Prod"
	Environment_Stage provider.EnvironmentZone = "Stag"
)

// Defines values for Region.
const (
	Region_Flex = "Flex"
	Region_IM   = "IM"
)

type Protocol string

const (
	Protocol_TCP = "TCP"
	Protocol_UDP = "UDP"
)

// ListResponse defines model for ListResponse.
type ListResponse struct {
	Result []interface{} `json:"result"`
}

// tokenResponse defines model for auth.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	accessToken string

	// The endpoint of the baseURL conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the baseURL, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the baseURL.
	baseURL string

	// Prod or Staging
	environment provider.EnvironmentZone

	// The region the operator is managing (i.e. Flex, IM, etc)
	region string

	configuration *config.Configuration

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	mu sync.Mutex
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// NewClient a new Client, with reasonable defaults
func NewClient(server, environment, region string, configuration *config.Configuration, opts ...ClientOption) (provider.FirewallProvider, error) {

	var env provider.EnvironmentZone
	switch strings.ToLower(environment) {
	case "prod":
		env = Environment_Prod
	case "stag":
		env = Environment_Stage
	default:
		return nil, fmt.Errorf("invalid environment: %s", environment)
	}

	// Handle casing of the region in the FW API. This is temporary until we move
	// to the new naming scehme.
	switch strings.ToLower(region) {
	case "flex":
		region = Region_Flex
	case "im":
		region = Region_IM
	}

	// create a client with sane default values
	client := Client{
		baseURL:       server,
		environment:   env,
		region:        region,
		configuration: configuration,
	}

	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the baseURL URL always has a trailing slash
	if !strings.HasSuffix(client.baseURL, "/") {
		client.baseURL += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// flattenCurrentRules takes a set of rules from the FW API and flattens them into a table of
// rules filtering out only those that match the vip==destIP
func (c *Client) flattenCurrentRules(ctx context.Context, response *CurrentRulesResult, vip string) []provider.Rule {
	log := log.FromContext(ctx)

	rules := []provider.Rule{}

	if response == nil {
		return rules
	}

	for _, entry := range response.CurrentRules {
		for _, sourceIp := range entry.SourceAddress {
			if strings.HasPrefix(sourceIp, "h_") || strings.HasPrefix(sourceIp, "n_") {
				continue
			}

			for _, destIp := range entry.DestAddress {
				if strings.HasPrefix(destIp, "h_") || strings.HasPrefix(destIp, "n_") || destIp != vip {
					continue
				}

				for _, port := range entry.Port {

					// Split the port into protocol / port
					splitPort := strings.Split(port, "_")

					// Verify the port is split into two parts, the protocol + port.
					// This is indexed later in the code and will cause a panic if not split correctly.
					if len(splitPort) != 2 {
						log.Error(fmt.Errorf("invalid port format: %s", port), "error flattening rules")
						continue
					}

					// If the source IP ends with a "/32", remove it. This can cause issues since
					// the FW API sometimes adds this to a rule and cuases the rules comparisons
					// to fail.
					if strings.HasSuffix(sourceIp, "/32") {
						sourceIp = strings.Replace(sourceIp, "/32", "", 1)
					}

					rules = append(rules, provider.Rule{
						DestIp:      destIp,
						Port:        splitPort[1],
						Protocol:    splitPort[0],
						SourceIp:    sourceIp,
						Environment: c.environment,
						Region:      c.region,
						CustomerId:  entry.CustomerId,
					})
				}
			}
		}
	}
	return rules
}

func (c *Client) flattenFWRules(fwRules []v1alpha1.FirewallRule, customerId string) []provider.Rule {
	rules := []provider.Rule{}

	for _, fw := range fwRules {
		for _, sourceIp := range fw.Spec.SourceIPs {

			rules = append(rules, provider.Rule{
				DestIp:      fw.Spec.DestinationIP,
				Port:        fw.Spec.Port,
				Protocol:    string(fw.Spec.Protocol),
				SourceIp:    sourceIp,
				CustomerId:  customerId,
				Region:      c.region,
				Environment: c.environment,
			})
		}
	}
	return rules
}

type RuleResult struct {
	CustomerId    string   `json:"customer_id"`
	Environment   string   `json:"environment"`
	Region        string   `json:"region"`
	RuleName      string   `json:"rule_name"`
	SourceAddress []string `json:"source_address"`
	DestAddress   []string `json:"dest_address"`
	Port          []string `json:"ports"`
	Protocol      string   `json:"protocol"`
}

type CurrentRulesResult struct {
	CurrentRules []RuleResult `json:"result"`
}

func (c *Client) SyncFirewallRules(ctx context.Context, desiredRules []v1alpha1.FirewallRule, existingRules []provider.Rule, vip, cloudAccountId string) error {

	log := log.FromContext(ctx)

	if len(desiredRules) == 0 {
		return fmt.Errorf("missing desired rules")
	}

	// Flatten the desired rules
	desired := c.flattenFWRules(desiredRules, cloudAccountId)

	// Determine which rules ar missing and should be created in the Firewall.
	rulesToAdd, rulesToRemove := calculateRulesToAddRemove(ctx, desired, existingRules, cloudAccountId)

	for rule, sourceIPs := range rulesToRemove {

		port, err := strconv.Atoi(rule.Port)
		if err != nil {
			return err
		}

		ruleExists, err := c.RuleExists(ctx, sourceIPs, rule.DestIp, rule.Protocol, strconv.Itoa(port))
		if err != nil {
			return err
		}

		if ruleExists {

			log.Info("removing existing fwrule",
				"sourceIP", strings.Join(sourceIPs, ", "),
				"destIP", rule.DestIp,
				"protocol", rule.Protocol,
				"port", rule.Port,
				"customerid", rule.CustomerId)

			_, err := c.removeAccess(ctx, provider.Rule{
				CustomerId:  rule.CustomerId,
				DestIp:      rule.DestIp,
				SourceIp:    strings.Join(sourceIPs, ", "),
				Port:        rule.Port,
				Protocol:    rule.Protocol,
				Region:      c.region,
				Environment: c.environment,
			})
			if err != nil {
				return err
			}
		}
	}

	for rule, sourceIPs := range rulesToAdd {

		// Rule did not exist, create it
		log.Info("creating missing fwrule",
			"sourceIP", strings.Join(sourceIPs, ","),
			"destIP", rule.DestIp,
			"protocol", rule.Protocol,
			"port", rule.Port,
			"customerid", cloudAccountId)

		_, err := c.requestPorts(ctx, provider.Rule{
			CustomerId:  cloudAccountId,
			DestIp:      rule.DestIp,
			SourceIp:    strings.Join(sourceIPs, ","),
			Port:        rule.Port,
			Protocol:    rule.Protocol,
			Region:      c.region,
			Environment: c.environment,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type RuleRequest struct {
	DestIp     string `json:"dest_ip"`
	Port       string `json:"port"`
	Protocol   string `json:"protocol"`
	CustomerId string `json:"customer_id"`
}

type SourceIPs []string

func calculateRulesToAddRemove(ctx context.Context, desired []provider.Rule, existing []provider.Rule, cloudAccountId string) (map[RuleRequest]SourceIPs, map[RuleRequest]SourceIPs) {

	log := log.FromContext(ctx)

	rulesToAdd := make(map[RuleRequest]SourceIPs)
	rulesToRemove := make(map[RuleRequest]SourceIPs)

	// Compare the two sets of rules
	for _, desiredRule := range desired {

		// Check if the rule exists already
		if index, found := ruleExists(existing, desiredRule); found {
			// Remove the rule since it was found to be existing, no need
			// to further process
			log.Info("fwrule exists, no change required",
				"sourceIP", desiredRule.SourceIp,
				"destIP", desiredRule.DestIp,
				"protocol", desiredRule.Protocol,
				"port", desiredRule.Port,
				"customerid", cloudAccountId)

			existing = removeRule(existing, index)
		} else {

			// Add the rule since it was not found in the existing set
			log.Info("fwrule missing, adding",
				"sourceIP", desiredRule.SourceIp,
				"destIP", desiredRule.DestIp,
				"protocol", desiredRule.Protocol,
				"port", desiredRule.Port,
				"customerid", cloudAccountId)

			req := RuleRequest{
				DestIp:     desiredRule.DestIp,
				Port:       desiredRule.Port,
				Protocol:   desiredRule.Protocol,
				CustomerId: cloudAccountId,
			}

			sourceIPs, found := rulesToAdd[req]
			if !found {
				rulesToAdd[req] = []string{desiredRule.SourceIp}
				continue
			}

			rulesToAdd[req] = append(sourceIPs, desiredRule.SourceIp)
		}
	}

	// Any rule left in existing should be removed
	for _, e := range existing {

		// Add the rule since it was not found in the existing set
		//
		// NOTE: Important to use the customerId from the rule since it
		// potentially cloud be different than the current customer.
		log.Info("fwrule not needed, remove",
			"sourceIP", e.SourceIp,
			"destIP", e.DestIp,
			"protocol", e.Protocol,
			"port", e.Port,
			"customerid", e.CustomerId)

		req := RuleRequest{
			DestIp:     e.DestIp,
			Port:       e.Port,
			Protocol:   e.Protocol,
			CustomerId: e.CustomerId,
		}

		sourceIPs, found := rulesToRemove[req]
		if !found {
			rulesToRemove[req] = []string{e.SourceIp}
			continue
		}

		rulesToRemove[req] = append(sourceIPs, e.SourceIp)
	}

	return rulesToAdd, rulesToRemove
}

// ruleExists determines is the rule exists in the set of rules passed in.
// Returns true if it exists along with the index of the found item in the existing
// set passed.
func ruleExists(existingRules []provider.Rule, desiredRule provider.Rule) (int, bool) {
	for i, existingRule := range existingRules {
		if desiredRule.DestIp == existingRule.DestIp &&
			desiredRule.Port == existingRule.Port &&
			strings.EqualFold(desiredRule.Protocol, existingRule.Protocol) &&
			desiredRule.SourceIp == existingRule.SourceIp {
			return i, true
		}
	}
	return -1, false
}

func removeRule(rules []provider.Rule, index int) []provider.Rule {
	return append(rules[:index], rules[index+1:]...)
}

func (c *Client) getExistingCustomerAccessSource(ctx context.Context, customerId string, vip string) (*CurrentRulesResult, error) {
	return c.getAccess(ctx, customerId, vip)
}

func (c *Client) GetExistingCustomerAccess(ctx context.Context, customerId string, vip string) ([]provider.Rule, error) {

	response, err := c.getAccess(ctx, customerId, vip)
	if err != nil {
		return nil, err
	}

	return c.flattenCurrentRules(ctx, response, vip), nil
}

func (c *Client) getAccess(ctx context.Context, customerId string, vip string) (*CurrentRulesResult, error) {
	var err error

	serverURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	operationPath := "/api/getallaccess"
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()
	queryValues.Add("environment", string(c.environment))
	queryValues.Add("region", c.region)
	queryValues.Add("vip", vip)
	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	token, err := c.requestToken(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error requesting existing vip access")
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// If there aren't any rules, don't try to unmarshal the response and return empty set.
	if strings.Contains(string(body), "No existing firewall rules found for vip") || strings.Contains(string(body), "No existing IDCAPI firewall rules found") {
		return nil, nil
	}

	var response CurrentRulesResult
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) RuleExists(ctx context.Context, sourceIPs []string, dest, protocol, port string) (bool, error) {
	var err error

	serverURL, err := url.Parse(c.baseURL)
	if err != nil {
		return false, err
	}

	operationPath := "/api/ports"
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return false, err
	}

	queryValues := queryURL.Query()

	queryValues.Add("source_ip", strings.Join(sourceIPs, ","))
	queryValues.Add("environment", string(c.environment))
	queryValues.Add("region", string(c.region))
	queryValues.Add("dest_ip", dest)
	queryValues.Add("protocol", protocol)
	queryValues.Add("port", port)
	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return false, err
	}

	token, err := c.requestToken(ctx)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	req = req.WithContext(ctx)
	res, err := c.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	if res.StatusCode != 200 {
		// Ignore 500 errors
		return false, nil //fmt.Errorf("non 200 status code returned, got: %d", res.StatusCode)
	}

	var response provider.RequestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return false, err
	}

	if response.Result == "One or more of the requested flows is already allowed." {
		return true, nil
	}

	return false, nil
}

func (c *Client) requestPorts(ctx context.Context, rule provider.Rule) (*provider.RequestResponse, error) {
	var err error

	log := log.FromContext(ctx)

	serverURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	operationPath := "/api/ports"
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	// Convert the Rule into a JSON payload
	fwRule := provider.Rule{
		CustomerId:  rule.CustomerId,
		SourceIp:    rule.SourceIp,
		Environment: c.environment,
		Region:      c.region,
		DestIp:      rule.DestIp,
		Protocol:    rule.Protocol,
		Port:        rule.Port,
	}

	buf, err := json.Marshal(fwRule)
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("requestPorts: %s %s", queryURL, string(buf)))

	req, err := http.NewRequest("POST", queryURL.String(), bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	token, err := c.requestToken(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("firewall api returned an error: %s", body)
	}

	var response provider.RequestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	// Check the response message, if one or more source IPs are already allowed, the FW API
	// will return an error message with a http 200.
	// If this is the case, return an error because the operator did not function as intended even
	// though it returned an HTTP 200 response.
	if response.Result == "One or more of the requested flows is already allowed." {
		return nil, fmt.Errorf("one or more rules already exist")
	}

	return &provider.RequestResponse{
		Result: string(body),
	}, nil
}

// RemoveAccess removes a firewall rule from the firewall based on the provided FirewallRule.
// It first fetches the existing rules from the firewall to ensure it removes the correct rule,
// even if the rule has been changed in the CRD but not yet applied to the firewall.
//
// Parameters:
//   - ctx: The context for the request.
//   - fwRule: The firewall rule to be removed.
//
// Returns:
//   - *provider.RequestResponse: The response indicating the result of the request.
//   - error: An error if the operation fails.
func (c *Client) RemoveAccess(ctx context.Context, fwRule v1alpha1.FirewallRule) (*provider.RequestResponse, error) {

	log := log.FromContext(ctx)

	cloudAccountId, err := provider.GetCloudAccountId(fwRule)
	if err != nil {
		return nil, err
	}

	// To delete the rule, first get the access that exists in the FW. This is important because it's
	// possible that the rule has has been changed in the CRD, but not been applied to the FW yet.
	// In this case, we need to fetch what actually exists in the FW then remove those rules.
	rules, err := c.getExistingCustomerAccessSource(ctx, cloudAccountId, fwRule.Spec.DestinationIP)
	if err != nil {
		return nil, err
	}

	// If there are no rules, then there is nothing to remove.
	// This is a valid case and should not be considered an error.
	if rules == nil {
		log.Info("no rules found to remove")
		return &provider.RequestResponse{Result: "success: no rules to remove"}, nil
	}

	// Find the rule which matches the port on this rule
	for _, rule := range rules.CurrentRules {

		// Iterate over the response ports and find the one that matches the port on the rule.
		// It's possible the rules are combined in the FW and multiple ports exist for this rule.
		for _, port := range rule.Port {

			// Split the port into protocol / port
			splitPort := strings.Split(port, "_")

			// Verify the port is split into two parts, the protocol + port.
			// This is indexed later in the code and will cause a panic if not split correctly.
			if len(splitPort) != 2 {
				log.Error(fmt.Errorf("invalid port format: %s", port), "error removing rule")
				continue
			}

			if splitPort[1] == fwRule.Spec.Port {

				// Remove the rule
				_, err := c.removeAccess(ctx, provider.Rule{
					CustomerId:  rule.CustomerId,
					SourceIp:    strings.Join(rule.SourceAddress, ","),
					Environment: c.environment,
					Region:      c.region,
					DestIp:      fwRule.Spec.DestinationIP,
					Protocol:    splitPort[0],
					Port:        splitPort[1],
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &provider.RequestResponse{Result: "success"}, nil
}

func (c *Client) removeAccess(ctx context.Context, rule provider.Rule) (*provider.RequestResponse, error) {
	var err error
	log := log.FromContext(ctx)

	serverURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	operationPath := "/api/removeaccess"
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	// Convert the Rule into a JSON payload
	fwRule := provider.Rule{
		CustomerId:  rule.CustomerId,
		SourceIp:    rule.SourceIp,
		Environment: c.environment,
		Region:      c.region,
		DestIp:      rule.DestIp,
		Protocol:    rule.Protocol,
		Port:        rule.Port,
	}

	buf, err := json.Marshal(fwRule)
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("removeAccess: %s %s", queryURL, string(buf)))

	req, err := http.NewRequest("POST", queryURL.String(), bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	req = req.WithContext(ctx)

	token, err := c.requestToken(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error removing rule: %s", body)
	}

	return &provider.RequestResponse{
		Result: string(body),
	}, nil
}

func (c *Client) RemoveAccessPerCustomerId(ctx context.Context, customerId string) (*http.Response, error) {
	var err error

	serverURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}

	operationPath := "/api/removeallaccess"
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()
	queryValues.Add("customer_id", customerId)
	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	token, err := c.requestToken(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

// RequestToken takes a username/password and requests a new auth token.
func (c *Client) requestToken(ctx context.Context) (string, error) {

	// Making sure once process can check and update the token at a time
	c.mu.Lock()
	defer c.mu.Unlock()

	log := log.FromContext(ctx)

	// Check if the current auth token is valid, if so re-use
	valid, err := c.validToken(ctx, c.accessToken)
	if err != nil {
		return "", err
	}

	if valid {
		log.Info("auth token is valid, not requesting a new one")
		return c.accessToken, nil
	}

	serverURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	operationPath := "/token"
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return "", err
	}

	username, err := c.configuration.GetAPIUsername()
	if err != nil {
		return "", err
	}

	password, err := c.configuration.GetAPIPassword()
	if err != nil {
		return "", err
	}

	userAuth := UserSchema{
		Username: username,
		Password: password,
	}

	buf, err := json.Marshal(userAuth)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", queryURL.String(), bytes.NewBuffer(buf))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	res, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("invalid credentials")
	}

	var response tokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	// Persist auth token for future requests
	c.accessToken = response.AccessToken

	return response.AccessToken, nil
}

func (c *Client) validToken(ctx context.Context, token string) (bool, error) {
	var err error

	if token == "" {
		return false, nil
	}

	serverURL, err := url.Parse(c.baseURL)
	if err != nil {
		return false, err
	}

	operationPath := "/validate"
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	req = req.WithContext(ctx)
	res, err := c.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var response provider.RequestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return false, err
	}

	if response.Result == "Your token is valid." {
		return true, nil
	}

	return false, nil
}
