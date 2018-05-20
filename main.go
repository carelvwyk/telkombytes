package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"telkom/netclient"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

var (
	username   = flag.String("username", "", "Telkom username")
	password   = flag.String("password", "", "Telkom password")
	localDebug = flag.Bool("localdebug", false,
		"Run locally instead of lambda handler")
)

func main() {
	flag.Parse()

	if *localDebug {
		services, err := netclient.GetServiceBundles(*username, *password)
		if err != nil {
			log.Fatalf(err.Error())
		}
		for _, s := range services {
			log.Println(s)
		}
		return
	}

	lambda.Start(handler)
}

func handler() error {
	telkomUsernames := os.Getenv("TELKOM_USERNAMES")
	telkomPasswords := os.Getenv("TELKOM_PASSWORDS")

	if telkomUsernames == "" || telkomPasswords == "" {
		return fmt.Errorf("Must set env variables TELKOM_USERNAMES and TELKOM_PASSWORDS")
	}

	// Retrieve service and bundle information:
	users := strings.Split(telkomUsernames, ";")
	passwords := strings.Split(telkomPasswords, ";")
	services := []netclient.Service{}
	for i, user := range users {
		s, err := netclient.GetServiceBundles(user, passwords[i])
		if err != nil {
			return err
		}
		services = append(services, s...)
	}

	// Publish to aws cloudwatch metrics
	sess := session.Must(session.NewSession())

	return putTelkomMetricsToCloudwatch(sess, services)
}

func putTelkomMetricsToCloudwatch(sess *session.Session, services []netclient.Service) error {
	svc := cloudwatch.New(sess)

	for _, service := range services {
		_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
			Namespace: aws.String("TelkomBytes"),
			MetricData: []*cloudwatch.MetricDatum{
				&cloudwatch.MetricDatum{
					MetricName: aws.String("Standard Remaining"),
					Unit:       aws.String("Bytes"),
					Value:      aws.Float64(float64(service.NonFreeBytesRemaining())),
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("MSISDN"),
							Value: aws.String(service.Msisdn),
						},
					},
				},
				&cloudwatch.MetricDatum{
					MetricName: aws.String("NighSurfer Remaining"),
					Unit:       aws.String("Bytes"),
					Value:      aws.Float64(float64(service.NightSurferRemaining())),
					Dimensions: []*cloudwatch.Dimension{
						&cloudwatch.Dimension{
							Name:  aws.String("MSISDN"),
							Value: aws.String(service.Msisdn),
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
