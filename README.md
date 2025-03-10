# AWS SSO CLI
![Tests](https://github.com/synfinatic/aws-sso-cli/workflows/Tests/badge.svg)
[![Report Card Badge](https://goreportcard.com/badge/github.com/synfinatic/aws-sso-cli)](https://goreportcard.com/report/github.com/synfinatic/aws-sso-cli)
[![License Badge](https://img.shields.io/badge/license-GPLv3-blue.svg)](https://raw.githubusercontent.com/synfinatic/aws-sso-cli/main/LICENSE)
[![Codecov Badge](https://codecov.io/gh/synfinatic/aws-sso-cli/branch/main/graph/badge.svg?token=F8454GS4HS)](https://codecov.io/gh/synfinatic/aws-sso-cli)

 * [About](#about)
 * [What does AWS SSO CLI do?](#what-does-aws-sso-cli-do)
 * [Demo](#demo)
 * [Security](#security)
 * [Commands](#commands)
    * [cache](#cache)
    * [console](#console)
	* [config](#config)
	* [eval](#eval)
	* [exec](#exec)
	* [flush](#flush)
	* [list](#list)
	* [process](#process)
	* [tags](#tags)
	* [time](#time)
	* [install-completions](#install-completions)
 * [Environment Variables](#environment-variables)
 * [License](#license)

Other Pages:

 * [Quick Start & Installation Guide](docs/quickstart.md)
 * [Configuration](docs/config.md)
 * [Frequently Asked Questions](docs/FAQ.md)
 * [Compared to AWS Vault](docs/aws-vault.md)
 * [Releases](https://github.com/synfinatic/aws-sso-cli/releases)
 * [Changelog](CHANGELOG.md)


## About

AWS SSO CLI is a secure replacement for using the [aws configure sso](
https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html)
wizard with a focus on security and ease of use for organizations with
many AWS Accounts and/or users with many IAM Roles to assume. It shares
a lot in common with [aws-vault](https://github.com/99designs/aws-vault),
but is more focused on the AWS SSO use case instead of static API credentials.
Check out [this page](docs/aws-vault.md) for more information on how these
two tools compare.

AWS SSO CLI requires your AWS account(s) to be setup with [AWS SSO](
https://aws.amazon.com/single-sign-on/)!  If your organization is using the
older SAML integration (typically you will have multiple tiles in OneLogin/Okta)
then this won't work for you.

## What does AWS SSO CLI do?

AWS SSO CLI makes it easy to manage your shell environment variables allowing
you to access the AWS API using CLI tools.  Unlike the official AWS tooling,
the `aws-sso` command does not require defining named profiles in your
`~/.aws/config` (or anywhere else for that matter) for each and every role you
wish to assume and use.

Instead, it focuses on making it easy to select a role via CLI arguments or
via an interactive auto-complete experience with automatic and user-defined
metadata (tags) and exports the necessary [AWS STS Token credentials](
https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_use-resources.html#using-temp-creds-sdk-cli)
to your shell environment.

## Demo

Here's a quick demo showing how to select a role to assume in interactive mode
and then run commands in that context (by default it starts a new shell).

<!-- exec -->
[![asciicast](https://asciinema.org/a/462167.svg)](https://asciinema.org/a/462167)

Want to see more?  Check out the [other demos](docs/demos.md).

## Security

Unlike the official [AWS cli tooling](https://aws.amazon.com/cli/), _all_
authentication tokens and credentials used for accessing AWS and your SSO
provider are encrypted on disk using your choice of secure storage solution.
All encryption is handled by the [99designs/keyring](https://github.com/99designs/keyring)
library which is also used by [aws-vault](https://github.com/99designs/aws-vault).

Credentials encrypted by `aws-sso` and not via the standard AWS CLI tool:

 * AWS SSO ClientID/ClientSecret -- `~/.aws/sso/cache/botocore-client-id-<region>.json`
 * AWS SSO AccessToken -- `~/.aws/sso/cache/<random>.json`
 * AWS Profile Access Credentials -- `~/.aws/cli/cache/<random>.json`

As you can see, not only does the standard AWS CLI tool expose the temporary
AWS access credentials to your IAM roles, but more importantly the SSO
AccessToken which can be used to fetch IAM credentials for any role you have
been granted access!

### What is not encrypted?

 * Contents of user defined `~/.aws-sso/config.yaml`
 * Meta data associated with the AWS Roles fetched via AWS SSO in `~/.aws-sso/cache.json`
    * Email address tied to the account (root user)
    * AWS Account Alias
    * AWS Role ARN

## Commands

 * [cache](#cache) -- Force refresh of AWS SSO role information
 * [console](#console) -- Open AWS Console in a browser with the selected role
 * [config](#config) -- Update your `~/.aws/config` file with the AWS profiles in AWS SSO
 * [eval](#eval) -- Print shell environment variables for use in your shell
 * [exec](#exec) -- Exec a command with the selected role
 * [flush](#flush) -- Force delete of cached AWS SSO credentials
 * [list](#list) -- List all accounts & roles
 * [process](#process) -- Generate JSON for AWS profile credential\_process option
 * [tags](#tags) -- List manually created tags for each role
 * [time](#time) -- Print how much time remains for currently selected role
 * [install-completions](#install-completions) -- Install auto-complete functionality into your shell
 * `version` -- Print the version of aws-sso

### Common Flags

 * `--help`, `-h` -- Builtin and context sensitive help
 * `--browser <path>`, `-b` -- Override default browser to open AWS SSO URL (`$AWS_SSO_BROWSER`)
 * `--config <file>` -- Specify alternative config file (`$AWS_SSO_CONFIG`)
 * `--level <level>`, `-L` -- Change default log level: [error|warn|info|debug|trace]
 * `--lines` -- Print file number with logs
 * `--url-action`, `-u` -- How to handle URLs for your SSO provider
 * `--sso <name>`, `-S` -- Specify non-default AWS SSO instance to use (`$AWS_SSO`)
 * `--sts-refresh` -- Force refresh of STS Token Credentials

### console

Console generates a URL which will grant you access to the AWS Console in your
web browser.  The URL can be sent directly to the browser (default), printed
in the terminal or copied into the Copy & Paste buffer of your computer.

Flags:

 * `--region <region>`, `-r` -- Specify the `$AWS_DEFAULT_REGION` to use
 * `--arn <arn>`, `-a` -- ARN of role to assume (`$AWS_SSO_ROLE_ARN`)
 * `--account <account>`, `-A` -- AWS AccountID of role to assume (`$AWS_SSO_ACCOUNT_ID`)
 * `--duration <minutes>`, `-d` -- AWS Session duration in minutes (default 60)
 * `--prompt`, `-P` -- Force interactive prompt to select role
 * `--role <role>`, `-R` -- Name of AWS Role to assume (requires `--account`) (`$AWS_SSO_ROLE_NAME`)
 * `--profile <profile>`, `-p` -- Name of AWS Profile to assume

The generated URL is good for 15 minutes after it is created.

The common flag `--url-action` is used both for AWS SSO authentication as well as
what to do with the resulting URL from the `console` command.

Priority is given to:

 * `--prompt`
 * `--profile`
 * `--arn` (`$AWS_SSO_ROLE_ARN`)
 * `--account` (`$AWS_SSO_ACCOUNT_ID`) and `--role` (`$AWS_SSO_ROLE_NAME`)
 * `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN` environment variables
 * `AWS_PROFILE` environment variable (works with both SSO and static profiles)
 * Prompt user interactively

### config

Modifies the `~/.aws/config` file to contain a profile for every role accessible
via AWS SSO CLI.

Flags:

 * `--diff` -- Print a diff of changes to the config file instead of modifying it
 * `--open` -- Override how to open URls: [clip|exec|open] (required)
 * `--print` -- Print profile entries instead of modifying config file

This generates a series of [named profile entries](
https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html)
in the `~/.aws/config` file which allows you to easily use any AWS SSO role
just by setting the `$AWS_PROFILE` environment variable.  By default, each
profile is named according to the [ProfileFormat](docs/config.md#profileformat)
onfig option or overridden by the user defined [Profile](docs/config.md#profile)
option on a role by role basis.

For each profile generated, it will specify a [list of settings](
https://docs.aws.amazon.com/sdkref/latest/guide/settings-global.html) as defined
by the [ConfigVariables](docs/config.md#configvariables) setting in the
`~/.aws-sso/config.yaml`.

Unlike with other ways to use AWS SSO CLI, the AWS IAM STS credentials will
_automatically refresh_.  This means, if you do not have a valid AWS SSO token,
you will be prompted to authentiate via your SSO provider and subsequent
requests to obtain new IAM STS credentials will automatically happen as needed.

**Note:** Due to a limitation in the AWS tooling, `print` and `printurl` arn not
supported with `--url-action` when using the `$AWS_PROFILE` variable with AWS
SSO CLI.  Hence, you must use `open` or `exec` to auto-open URLs in your browser
(recommended) or `clip` to automatically copy URLs to your clipboard.

**Note:** You should run this command any time your list of AWS roles changes
in order to update the `~/.aws/config` file.

**Note:** It is important that you do _NOT_ remove the `# BEGIN_AWS_SSO_CLI` and
`# END_AWS_SSO_CLI` lines from your config file!  These markers are used to track
which profiles are managed by AWS SSO CLI.

**Note:** This command does not honor the `--sso` option as it operates on all
of the configured AWS SSO instances in the `~/.aws-sso/config.yaml` file.

### eval

Generate a series of `export VARIABLE=VALUE` lines suitable for sourcing into your
shell.  Allows obtaining new AWS credentials without starting a new shell.  Can be
used to refresh existing AWS credentials or by specifying the appropriate arguments.

Suggested use (bash): `eval $(aws-sso eval <args>)`

Flags:

 * `--arn <arn>`, `-a` -- ARN of role to assume
 * `--account <account>`, `-A` -- AWS AccountID of role to assume (requires `--role`)
 * `--role <role>`, `-R` -- Name of AWS Role to assume (requires `--account`)
 * `--profile <profile>`, `-p` -- Name of AWS Profile to assume
 * `--no-region` -- Do not set the AWS_DEFAULT_REGION from config.yaml
 * `--refresh` -- Refresh current IAM credentials

Priority is given to:

 * `--refresh` (Uses `$AWS_SSO_ROLE_ARN`)
 * `--profile`
 * `--arn`
 * `--account` and `--role`

**Note:** The `eval` command only honors the `$AWS_SSO_ROLE_ARN` in the context
of the `--refresh` flag.  The `$AWS_SSO_ROLE_NAME` and `$AWS_SSO_ACCOUNT_ID`
are always ignored.

**Note:** Using `--url-action=print` is supported, but you must be able to see the output
of _STDERR_ to see the URL to open.

**Note:** The `eval` command is not supported under Windows CommandPrompt or PowerShell.

See [Environment Variables](#environment-variables) for more information about
what varibles are set.

### exec

Exec allows you to execute a command with the necessary [AWS environment variables](
https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html).  By default,
if no command is specified, it will start a new interactive shell so you can run multiple
commands.

Flags:

 * `--arn <arn>`, `-a` -- ARN of role to assume (`$AWS_SSO_ROLE_ARN`)
 * `--account <account>`, `-A` -- AWS AccountID of role to assume (`$AWS_SSO_ACCOUNT_ID`)
 * `--env`, `-e` -- Use existing ENV vars generated by AWS SSO to generate a URL
 * `--role <role>`, `-R` -- Name of AWS Role to assume (`$AWS_SSO_ROLE_NAME`)
 * `--profile <profile>`, `-p` -- Name of AWS Profile to assume
 * `--no-region` -- Do not set the AWS_DEFAULT_REGION from config.yaml

Arguments: `[<command>] [<args> ...]`

Priority is given to:

 * `--profile`
 * `--arn` (`$AWS_SSO_ROLE_ARN`)
 * `--account` (`$AWS_SSO_ACCOUNT_ID`) and `--role` (`$AWS_SSO_ROLE_NAME`)
 * Prompt user interactively

You can not run `exec` inside of another `exec` shell.

See [Environment Variables](#environment-variables) for more information about what varibles are set.

### process

Process allows you to use AWS SSO as an [external credentials provider](
https://docs.aws.amazon.com/cli/latest/topic/config-vars.html#sourcing-credentials-from-external-processes)
with profiles defined in `~/.aws/config`.

Flags:

 * `--arn <arn>`, `-a` -- ARN of role to assume
 * `--account <account>`, `-A` -- AWS AccountID of role to assume
 * `--role <role>`, `-R` -- Name of AWS Role to assume (requires `--account`)
 * `--profile <profile>`, `-p` -- Name of AWS Profile to assume

Priority is given to:

 * `--profile`
 * `--arn`
 * `--account` and `--role`

**Note:** The `process` command does not honor the `$AWS_SSO_ROLE_ARN`, `$AWS_SSO_ACCOUNT_ID`, or
`$AWS_SSO_ROLE_NAME` environment variables.

**Note:** Due to a limitation of the AWS tooling, setting `--url-action print` will cause an error
because of a limitation of the AWS tooling which prevents it from working.

### cache

AWS SSO CLI caches information about your AWS Accounts, Roles and Tags for better
perfomance.  By default it will refresh this information after 24 hours, but you
can force this data to be refreshed immediately.

Cache data is also automatically updated anytime the `config.yaml` file is modified.

### list

List will list all of the AWS Roles you can assume with the metadata/tags available
to be used for interactive selection with `exec`.  You can control which fields are
printed by specifying the field names as arguments.

Flags:

 * `--list-fields`, `-f` -- List the available fields to print

Arguments: `[<field> ...]`

The arguments are a list of fields to display in the report.  Overrides the
defaults and/or the specified `ListFields` in the `config.yaml`.

Default fields:

 * `AccountId`
 * `AccountAlias`
 * `RoleName`
 * `ExpiresStr`

### flush

Flush any cached AWS SSO/STS credentials.  By default, it only flushes the
temporary STS IAM role credentials for the selected SSO instance.

Flags:

 * `--type`, `-t` -- Type of credentials to flush:
    * `sts` -- Flush temporary STS credentials for IAM roles
    * `sso` -- Flush temporary AWS SSO credentials
	* `all` -- Flush temporary STS and SSO  credentials

### tags

Tags dumps a list of AWS SSO roles with the available metadata tags.

Flags:

 * `--account <account>` -- Filter results by AccountId
 * `--role <role>` -- Filter results by Role Name

By default the following key/values are available as tags to your roles:

 * `AccountID` -- AWS Account ID
 * `Role` -- AWS Role Name
 * `Email` -- Email address of root account associated with the AWS Account
 * `AccountName` -- Account Name for any role defined in config (see below)
 * `AccountAlias` --- AWS Account Alias defined by account administrator
 * `History` -- Tag tracking if this role was recently used.  See `HistoryLimit`
                in config.

### time

Print a string containing the number of hours and minutes that the current
AWS Role's STS credentials are valid for in the format of `HHhMMm`

**Note:** This command is only useful when you have STS credentials configured
in your shell via [eval](#eval) or [exec](#exec).

### install-completions

Configures your appropriate shell configuration file to add auto-complete
functionality for commands, flags and options.  Must restart your shell
for this to take effect.

Modifies the following file based on your shell:
 * `~/.bash_profile` -- bash
 * `~/.zshrc` -- zsh

## Environment Variables

### Honored Variables

The following environment variables are honored by `aws-sso`:

 * `AWS_SSO_FILE_PASSWORD` -- Password to use with the `file` SecureStore
 * `AWS_SSO_CONFIG` -- Specify an alternate path to the `aws-sso` config file
 * `AWS_SSO_BROWSER` -- Override default browser for AWS SSO login
 * `AWS_SSO` -- Override default AWS SSO instance to use
 * `AWS_SSO_ROLE_NAME` -- Used for `--role`/`-R` with some commands
 * `AWS_SSO_ACCOUNT_ID` -- Used for `--account`/`-A` with some commands
 * `AWS_SSO_ROLE_ARN` -- Used for `--arn`/`-a` with some commands and with `eval --refresh`

The `file` SecureStore will use the `AWS_SSO_FILE_PASSWORD` environment
variable for the password if it is set. (Not recommended.)

Additionally, `$AWS_PROFILE` is honored via the standard AWS tooling when using
the [config](#config) command to manage your `~/.aws/config` file.

### Managed Variables

The following [AWS environment variables](
https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html)
are automatically set by `aws-sso`:

 * `AWS_ACCESS_KEY_ID` -- Authentication identifier required by AWS
 * `AWS_SECRET_ACCESS_KEY` -- Authentication secret required by AWS
 * `AWS_SESSION_TOKEN` -- Authentication secret required by AWS
 * `AWS_DEFAULT_REGION` -- Region to use AWS with (will never override an existing value)

The following environment variables are specific to `aws-sso`:

 * `AWS_SSO_ACCOUNT_ID` -- The AccountID for your IAM role
 * `AWS_SSO_ROLE_NAME` -- The name of the IAM role
 * `AWS_SSO_ROLE_ARN` -- The full ARN of the IAM role
 * `AWS_SSO_SESSION_EXPIRATION`  -- The date and time when the IAM role credentials will expire
 * `AWS_SSO_DEFAULT_REGION` -- Tracking variable for `AWS_DEFAULT_REGION`
 * `AWS_SSO_PROFILE` -- User customizable varible using the [ProfileFormat](docs/config.md#profileformat) template
 * `AWS_SSO` -- AWS SSO instance name

## License

AWS SSO CLI is licnsed under the [GPLv3](LICENSE).
