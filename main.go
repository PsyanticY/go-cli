package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/urfave/cli"
)

func GetAccounts() map[string]string {
	accounts, err := os.Open("static/accounts.json")
	if err != nil {
		fmt.Println(err)
	}
	defer accounts.Close()
	var result map[string]string

	byteValue, _ := ioutil.ReadAll(accounts)
	json.Unmarshal(byteValue, &result)
	return result
}
func GetAccountId(account string) string {

	result := GetAccounts()
	return result[account]
}

func main() {
	app := cli.NewApp()
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "PsyanticY",
			Email: "iuns@outlook.fr",
		},
	}
	app.Copyright = "MIT PsyanticY (2019)"
	app.UsageText = "Let you query IPs, and other usefull stuff"
	app.Name = "CLI to run some useful commands"
	app.Usage = "Let you query IPs, Switch AWS Role and AWS account ID"
	app.Version = "0.0.1"
	myFlags := []cli.Flag{
		cli.StringFlag{
			Name:     "host",
			Required: true,
			Usage:    "Provide the hostname of the server to get it's IP",
		},
		cli.StringFlag{
			Name:   "domain, d",
			Value:  "",
			Usage:  "providing a domain name",
			EnvVar: "DOMAIN",
		},
	}
	swithcRoleFlags := []cli.Flag{
		cli.StringFlag{
			Required: true,
			Name:     "a, account",
			Value:    "",
			Usage:    "AWS account to switch role to",
		},
		cli.StringFlag{
			Required: true,
			Name:     "u, username",
			Value:    "",
			Usage:    "AWS Username.",
		},
		cli.StringFlag{
			Required: true,
			Name:     "mfa",
			Value:    "",
			Usage:    "AWS User mfa.",
		},
	}
	accountFlags := []cli.Flag{
		cli.StringFlag{
			Required: true,
			Name:     "a, account",
			Value:    "",
			Usage:    "AWS account",
		},
	}
	app.EnableBashCompletion = true
	// we create our commands
	app.Commands = []cli.Command{
		{
			Name:    "iplookup",
			Aliases: []string{"ip"},
			Usage:   "Looks up the IP addresses for a particular host",
			Flags:   myFlags,
			Action: func(c *cli.Context) error {
				hostname := c.String("host")
				domain := c.String("domain")
				if domain != "" {
					hostname = hostname + "." + domain
				}
				ip, err := net.LookupIP(hostname)
				if err != nil {
					fmt.Println(err)
				}
				// check for a better way to understand the version of the Ip Address
				for i := 0; i < len(ip); i++ {
					if ip[i].To4() == nil {
						result := fmt.Sprintf("IPV6 address of \"%s\" is:", hostname)
						fmt.Println(result, ip[i])
					} else if ip[i].To4() != nil {
						result := fmt.Sprintf("IPV4 address of \"%s\" is:", hostname)
						fmt.Println(result, ip[i])
					}
				}
				return nil
			},
		},
		{
			Name:    "switchrole",
			Aliases: []string{"sr"},
			Usage:   "Switch to another aws role ",
			Flags:   swithcRoleFlags,
			Action: func(c *cli.Context) error {
				username := c.String("username")
				account := c.String("account")
				mfa := c.String("mfa")
				role := "switchRoleRole"
				duration := int64(3600)
				accountNumber := GetAccountId(account)
				client, err := session.NewSession(&aws.Config{
					Region: aws.String("us-east-1")},
				)

				// Create a sts service client.
				svc := sts.New(client)
				roleArn := "arn:aws:iam::" + accountNumber + ":role/" + role
				mfaARN := "arn:aws:iam::211212121212:mfa/" + username

				input := &sts.AssumeRoleInput{
					DurationSeconds: aws.Int64(duration),
					RoleArn:         aws.String(roleArn),
					RoleSessionName: aws.String(account),
					TokenCode:       aws.String(mfa),
					SerialNumber:    aws.String(mfaARN),
				}

				result, err := svc.AssumeRole(input)
				if err != nil {
					if aerr, ok := err.(awserr.Error); ok {
						switch aerr.Code() {
						case sts.ErrCodeMalformedPolicyDocumentException:
							fmt.Println(sts.ErrCodeMalformedPolicyDocumentException, aerr.Error())
						case sts.ErrCodePackedPolicyTooLargeException:
							fmt.Println(sts.ErrCodePackedPolicyTooLargeException, aerr.Error())
						case sts.ErrCodeRegionDisabledException:
							fmt.Println(sts.ErrCodeRegionDisabledException, aerr.Error())
						default:
							fmt.Println(aerr.Error())
						}
					} else {
						// Print the error, cast err to awserr.Error to get the Code and
						// Message from an error.
						fmt.Println(err.Error())
					}
					return err
				}
				var creds credentials.Value
				outputRoleArn := *result.AssumedRoleUser.Arn
				// roleAssumedRoleId := *result.AssumedRoleUser.AssumedRoleId
				creds.AccessKeyID = *result.Credentials.AccessKeyId
				creds.SecretAccessKey = *result.Credentials.SecretAccessKey
				creds.SessionToken = *result.Credentials.SessionToken
				fmt.Println("--------------------------------------------------------------")
				fmt.Println("--------------------------------------------------------------")
				fmt.Println("-- user:", username, "successfully assumed role to", account)
				fmt.Println("-- Assumed Role:", outputRoleArn)
				fmt.Println("-- Setting environment variable ...")
				fmt.Println("-- Since the env variable will be cleared out when the program exit")
				fmt.Println("-- Copy past the following ...")
				fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", creds.AccessKeyID)
				fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", creds.SecretAccessKey)
				fmt.Printf("export AWS_SESSION_TOKEN=%s\n", creds.SessionToken)
				fmt.Println("--------------------------------------------------------------")
				fmt.Println("--------------------------------------------------------------")
				return nil
			},
		},
		{
			Name:    "get-account-id",
			Aliases: []string{"id"},
			Usage:   "Looks up the ID of the provided aws account name",
			Flags:   accountFlags,
			Action: func(c *cli.Context) error {
				account := c.String("account")
				accounts := GetAccounts()
				fmt.Println("--------------------------------------------------------------")
				if GetAccountId(account) != "" {
					fmt.Println("-- Account ID of:", account, "is:", GetAccountId(account))
                                        fmt.Println("--------------------------------------------------------------")
					return nil
				}

				fmt.Println("-- Was not able to find `" + account + "` account in my current list")
				for key, value := range accounts {
					if strings.Contains(key, account) {
						fmt.Printf("-- This account name is similar to the one you did provide...\n")
						fmt.Printf("-- %s: %s\n", key, value)
					}
				}
				fmt.Println("--------------------------------------------------------------")
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
