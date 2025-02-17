/*
Copyright © 2021 Red Hat, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package login

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/openshift/pagerduty-short-circuiter/cmd/kite/teams"
	"github.com/openshift/pagerduty-short-circuiter/pkg/client"
	"github.com/openshift/pagerduty-short-circuiter/pkg/config"
	"github.com/openshift/pagerduty-short-circuiter/pkg/constants"
	"github.com/spf13/cobra"
)

var loginArgs struct {
	apiKey      string
	accessToken string
}

var Cmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the PagerDuty CLI",
	Long: `The kite login command logs a user into PagerDuty CLI given a valid API key is provided. 
	You will have to login only once, all the kite commands are then available even if the terminal restarts.`,
	Args: cobra.NoArgs,
	RunE: loginHandler,
}

func init() {

	Cmd.Flags().StringVar(
		&loginArgs.apiKey,
		"api-key",
		"",
		"Access API key/token generated from "+constants.APIKeyURL+"\nUse this option to overwrite the existing API key.",
	)
	Cmd.Flags().StringVar(
		&loginArgs.accessToken,
		"access-token",
		"",
		"GitHub Personal Access Token generated from "+constants.AccessTokenURL+"\nUse this option to overwrite the existing Access Token.",
	)
}

// loginHandler handles the login flow into kite.
func loginHandler(cmd *cobra.Command, args []string) error {

	var user string
	var pdClient client.PagerDutyClient

	// load the configuration info
	cfg, err := config.Load()

	// If no config file can be located
	// Or if the config file has errors
	// Or if this is the first time a user is trying to login
	// A new configuration struct is initialized on login
	if err != nil {
		cfg = new(config.Config)
	}

	// If the key arg is not empty
	if loginArgs.apiKey != "" {

		cfg.ApiKey = loginArgs.apiKey

		// Save the key in the config file
		err = config.Save(cfg)

		if err != nil {
			return err
		}
	}

	// API key is not found in the config file
	if len(cfg.ApiKey) == 0 {

		// Create a new API key and store it in the config file
		err = generateNewKey(cfg)

		if err != nil {
			return err
		}
	}

	// API key is not found in the config file
	if len(cfg.AccessToken) == 0 {

		// Create a new API key and store it in the config file
		err = generateNewAccessToken(cfg)

		if err != nil {
			return err
		}

		// Save the key in the config file
		err = config.Save(cfg)

		if err != nil {
			return err
		}
	}

	// Connect to PagerDuty API client
	pdClient, err = client.NewClient().Connect()

	if err != nil {
		return err
	}

	// Login using the API key in the configuration file
	user, err = Login(cfg.ApiKey, pdClient)

	if err != nil {
		return err
	}

	// Print login success message
	successMessage(user)

	// Check if user has selected a team
	if cfg.TeamID == "" {
		teamdID, name, err := teams.SelectTeam(pdClient, os.Stdin)

		if err != nil {
			return err
		}

		cfg.TeamID = teamdID
		cfg.Team = name
	}

	// Save the Team ID to the config file
	err = config.Save(cfg)

	if err != nil {
		return err
	}

	return nil
}

// generateNewKey prompts the user to create a new API key and saves it to the config file.
func generateNewKey(cfg *config.Config) (err error) {
	//prompts the user to generate an API Key
	fmt.Println("In order to login it is mandatory to provide an API key.\nThe recommended way is to generate an API key via: " + constants.APIKeyURL)

	//Takes standard input from the user and stores it in a variable
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("API Key: ")

	cfg.ApiKey, err = reader.ReadString('\n')

	if err != nil {
		return err
	}

	return nil
}

// generateNewKey prompts the user to create a new API key and saves it to the config file.
func generateNewAccessToken(cfg *config.Config) (err error) {
	//prompts the user to generate an API Key
	fmt.Println("\nIn order to view SOP it is mandatory to provide an GitHub Access Token.\nThe recommended way is to generate a token via: " + constants.AccessTokenURL)

	//Takes standard input from the user and stores it in a variable
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("GitHub Access Token: ")

	cfg.AccessToken, err = reader.ReadString('\n')
	cfg.AccessToken = strings.TrimSuffix(cfg.AccessToken, "\n")

	if err != nil {
		return err
	}

	return nil
}

// Login handles PagerDuty REST API authentication via an user API token.
// Requests that cannot be authenticated will return a `401 Unauthorized` error response.
// It returns the username of the currently logged in user.
func Login(apiKey string, client client.PagerDutyClient) (string, error) {

	user, err := client.GetCurrentUser(pagerduty.GetCurrentUserOptions{})

	if err != nil {
		var apiError pagerduty.APIError

		//`401 Unauthorized` error response
		if errors.As(err, &apiError) {
			err = fmt.Errorf("login failed\n%v Unauthorized", apiError.StatusCode)
			return "", err
		}

		return "", err
	}

	return user.Name, nil
}

// successMessage prints the currently logged in username to the console
// if pagerduty login is successful.
func successMessage(user string) {
	fmt.Printf("Successfully logged in as user: %s\n", user)
}
