package main

/*
 * AWS SSO CLI
 * Copyright (c) 2021-2022 Aaron Turner  <synfinatic at gmail dot com>
 *
 * This program is free software: you can redistribute it
 * and/or modify it under the terms of the GNU General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or with the authors permission any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/c-bata/go-prompt"
	"github.com/synfinatic/aws-sso-cli/sso"
	"github.com/synfinatic/aws-sso-cli/utils"
)

type ExecCmd struct {
	// AWS Params
	Arn       string `kong:"short='a',help='ARN of role to assume',env='AWS_SSO_ROLE_ARN',predictor='arn'"`
	AccountId int64  `kong:"name='account',short='A',help='AWS AccountID of role to assume',env='AWS_SSO_ACCOUNT_ID',predictor='accountId'"`
	Role      string `kong:"short='R',help='Name of AWS Role to assume',env='AWS_SSO_ROLE_NAME',predictor='role'"`
	Profile   string `kong:"short='p',help='Name of AWS Profile to assume',predictor='profile'"`
	NoRegion  bool   `kong:"short='n',help='Do not set AWS_DEFAULT_REGION from config.yaml'"`

	// Exec Params
	Cmd  string   `kong:"arg,optional,name='command',help='Command to execute',env='SHELL'"`
	Args []string `kong:"arg,optional,passthrough,name='args',help='Associated arguments for the command'"`
}

func (cc *ExecCmd) Run(ctx *RunContext) error {
	err := checkAwsEnvironment()
	if err != nil {
		log.WithError(err).Fatalf("Unable to continue")
	}

	if runtime.GOOS == "windows" && ctx.Cli.Exec.Cmd == "" {
		// Windows doesn't set $SHELL, so default to CommandPrompt
		ctx.Cli.Exec.Cmd = "cmd.exe"
	}

	// Did user specify the ARN or account/role?
	if ctx.Cli.Exec.Profile != "" {
		awssso := doAuth(ctx)
		cache := ctx.Settings.Cache.GetSSO()
		rFlat, err := cache.Roles.GetRoleByProfile(ctx.Cli.Exec.Profile, ctx.Settings)
		if err != nil {
			return err
		}

		return execCmd(ctx, awssso, rFlat.AccountId, rFlat.RoleName)
	} else if ctx.Cli.Exec.Arn != "" {
		awssso := doAuth(ctx)

		accountid, role, err := utils.ParseRoleARN(ctx.Cli.Exec.Arn)
		if err != nil {
			return err
		}

		return execCmd(ctx, awssso, accountid, role)
	} else if ctx.Cli.Exec.AccountId != 0 || ctx.Cli.Exec.Role != "" {
		if ctx.Cli.Exec.AccountId == 0 || ctx.Cli.Exec.Role == "" {
			return fmt.Errorf("Please specify both --account and --role")
		}
		awssso := doAuth(ctx)

		return execCmd(ctx, awssso, ctx.Cli.Exec.AccountId, ctx.Cli.Exec.Role)
	}

	// Nope, auto-complete mode...
	sso, err := ctx.Settings.GetSelectedSSO(ctx.Cli.SSO)
	if err != nil {
		return err
	}
	if err = ctx.Settings.Cache.Expired(sso); err != nil {
		log.Infof(err.Error())
		c := &CacheCmd{}
		if err = c.Run(ctx); err != nil {
			return err
		}
	}

	sso.Refresh(ctx.Settings)
	fmt.Printf("Please use `exit` or `Ctrl-D` to quit.\n")

	c := NewTagsCompleter(ctx, sso, execCmd)
	opts := ctx.Settings.DefaultOptions(c.ExitChecker)
	opts = append(opts, ctx.Settings.GetColorOptions()...)

	p := prompt.New(
		c.Executor,
		c.Complete,
		opts...,
	)
	p.Run()
	return nil
}

// Executes Cmd+Args in the context of the AWS Role creds
func execCmd(ctx *RunContext, awssso *sso.AWSSSO, accountid int64, role string) error {
	region := ctx.Settings.GetDefaultRegion(ctx.Cli.Exec.AccountId, ctx.Cli.Exec.Role, ctx.Cli.Exec.NoRegion)

	ctx.Settings.Cache.AddHistory(utils.MakeRoleARN(accountid, role))
	if err := ctx.Settings.Cache.Save(false); err != nil {
		log.WithError(err).Warnf("Unable to update cache")
	}

	// ready our command and connect everything up
	cmd := exec.Command(ctx.Cli.Exec.Cmd, ctx.Cli.Exec.Args...) // #nosec
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ() // copy our current environment to the executor

	// add the variables we need for AWS to the executor without polluting our
	// own process
	for k, v := range execShellEnvs(ctx, awssso, accountid, role, region) {
		log.Debugf("Setting %s = %s", k, v)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	// just do it!
	return cmd.Run()
}

func execShellEnvs(ctx *RunContext, awssso *sso.AWSSSO, accountid int64, role, region string) map[string]string {
	var err error
	credsPtr := GetRoleCredentials(ctx, awssso, accountid, role)
	creds := *credsPtr

	ssoName, _ := ctx.Settings.GetSelectedSSOName(ctx.Cli.SSO)
	shellVars := map[string]string{
		"AWS_ACCESS_KEY_ID":          creds.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY":      creds.SecretAccessKey,
		"AWS_SESSION_TOKEN":          creds.SessionToken,
		"AWS_SSO_ACCOUNT_ID":         creds.AccountIdStr(),
		"AWS_SSO_ROLE_NAME":          creds.RoleName,
		"AWS_SSO_SESSION_EXPIRATION": creds.ExpireString(),
		"AWS_SSO_ROLE_ARN":           utils.MakeRoleARN(creds.AccountId, creds.RoleName),
		"AWS_SSO":                    ssoName,
	}

	if len(region) > 0 {
		shellVars["AWS_DEFAULT_REGION"] = region
		shellVars["AWS_SSO_DEFAULT_REGION"] = region
	} else {
		// we no longer manage the region
		shellVars["AWS_SSO_DEFAULT_REGION"] = ""
	}

	// Set the AWS_SSO_PROFILE env var using our template
	cache := ctx.Settings.Cache.GetSSO()
	var roleInfo *sso.AWSRoleFlat
	if roleInfo, err = cache.Roles.GetRole(accountid, role); err != nil {
		// this error should never happen
		log.Errorf("Unable to find role in cache.  Unable to set AWS_SSO_PROFILE")
	} else {
		shellVars["AWS_SSO_PROFILE"], err = roleInfo.ProfileName(ctx.Settings)
		if err != nil {
			log.Errorf("Unable to generate AWS_SSO_PROFILE: %s", err.Error())
		}

		// and any EnvVarTags
		for k, v := range roleInfo.GetEnvVarTags(ctx.Settings) {
			shellVars[k] = v
		}
	}

	return shellVars
}

// returns an error if we have existing AWS env vars
func checkAwsEnvironment() error {
	checkVars := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_PROFILE"}
	for _, envVar := range checkVars {
		if _, ok := os.LookupEnv(envVar); ok {
			return fmt.Errorf("Conflicting environment variable '%s' is set", envVar)
		}
	}
	return nil
}
