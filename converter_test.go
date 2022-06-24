package rules2cron_test

import (
	"testing"
	"time"

	"github.com/mashiike/rules2cron"
	"github.com/stretchr/testify/require"
)

func TestConverter(t *testing.T) {
	defaultReferenceDate := time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		scheduleExpression string
		referenceDate      time.Time
		timeZone           *time.Location
		expectedCrontab    string
		expectedError      string
	}{
		//https://docs.aws.amazon.com/ja_jp/lambda/latest/dg/services-cloudwatchevents-expressions.html
		{
			scheduleExpression: "rate(1 minute)",
			expectedCrontab:    "* * * * *",
		},
		{
			scheduleExpression: "rate(2 minutes)",
			expectedCrontab:    "*/2 * * * *",
		},
		{
			scheduleExpression: "rate(1 hour)",
			expectedCrontab:    "0 * * * *",
		},
		{
			scheduleExpression: "rate(2 hours)",
			expectedCrontab:    "0 */2 * * *",
		},
		{
			scheduleExpression: "rate(1 day)",
			expectedCrontab:    "0 0 * * *",
		},
		{
			scheduleExpression: "rate(2 days)",
			expectedCrontab:    "0 0 */2 * *",
		},
		{
			scheduleExpression: "rate(1 day)",
			expectedCrontab:    "0 9 * * *",
			timeZone:           Must(time.LoadLocation("Asia/Tokyo")),
		},
		{
			scheduleExpression: "rate(1 day)",
			expectedCrontab:    "0 17 * * *",
			timeZone:           Must(time.LoadLocation("America/Los_Angeles")),
		},
		{
			scheduleExpression: "rate(1 days)",
			expectedError:      "invalid format: can not use pluralistic",
		},
		{
			scheduleExpression: "rate(2 days)",
			expectedCrontab:    "0 9 */2 * *",
			timeZone:           Must(time.LoadLocation("Asia/Tokyo")),
		},
		{
			scheduleExpression: "cron(15 10 * * ? *)",
			expectedCrontab:    "15 10 * * *",
		},
		{
			scheduleExpression: "cron(15 10-12/2 * * ? *)",
			expectedCrontab:    "15 10-12/2 * * *",
		},
		{
			scheduleExpression: "cron(15 10,11 * * ? *)",
			expectedCrontab:    "15 19,20 * * *",
			timeZone:           Must(time.LoadLocation("Asia/Tokyo")),
		},
		{
			scheduleExpression: "cron(15 10-12/2 * * ? *)",
			expectedCrontab:    "15 19-21/2 * * *",
			timeZone:           Must(time.LoadLocation("Asia/Tokyo")),
		},
		{
			scheduleExpression: "cron(15 * * 1 ? 2030-2040/2)",
			expectedError:      "cannot be converted because the reference date is not the target year: 2030-2040/2",
			referenceDate:      time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * * 1 ? 2020-2030/2)",
			expectedCrontab:    "15 * * 1 *",
			referenceDate:      time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * * 1 ? */2)",
			expectedCrontab:    "15 * * 1 *",
			referenceDate:      time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * * 1 ? */5)",
			expectedError:      "cannot be converted because the reference date is not the target year: */5",
			referenceDate:      time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * * * ? 2022)",
			expectedCrontab:    "15 * * * *",
			referenceDate:      time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * * * ? 2023)",
			expectedError:      "cannot be converted because the reference date is not the target year: 2023",
			referenceDate:      time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(0 18 ? * MON-FRI *)",
			expectedCrontab:    "0 18 * * 1-5",
		},
		{
			scheduleExpression: "cron(0 18 ? * SUN-WED *)",
			expectedCrontab:    "0 18 * * 0-3",
		},
		{
			scheduleExpression: "cron(0 18 ? * TUE-SUN *)",
			expectedCrontab:    "0 18 * * 2-7",
		},
		{
			scheduleExpression: "cron(0 18 ? * THU *)",
			expectedCrontab:    "0 18 * * 4",
		},
		{
			scheduleExpression: "cron(0 18 ? JAN THU *)",
			expectedCrontab:    "0 18 * 1 4",
		},
		{
			scheduleExpression: "cron(0/10 * ? * MON-FRI *)",
			expectedCrontab:    "0/10 * * * 1-5",
		},
		{
			scheduleExpression: "cron(0/10 * ? * 1-7 *)",
			expectedCrontab:    "0/10 * * * 0-6",
		},
		{
			scheduleExpression: "cron(15 * 2W * ? *)",
			expectedCrontab:    "15 * 1 * *",
			referenceDate:      time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * L * ? *)",
			expectedCrontab:    "15 * 31 * *",
			referenceDate:      time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * L * ? *)",
			expectedCrontab:    "15 * 30 * *",
			referenceDate:      time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * ? * 3L *)",
			expectedCrontab:    "15 * 28 * *",
			referenceDate:      time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			scheduleExpression: "cron(15 * ? * 3#2 *)",
			expectedCrontab:    "15 * 12 * *",
			referenceDate:      time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, c := range cases {
		t.Run(c.scheduleExpression, func(t *testing.T) {
			if c.referenceDate.IsZero() {
				c.referenceDate = defaultReferenceDate
			}
			if c.timeZone == nil {
				c.timeZone = time.UTC
			}
			converter := &rules2cron.Converter{
				ReferenceDate: c.referenceDate,
				TimeZone:      c.timeZone,
			}
			actual, err := converter.Convert(c.scheduleExpression)
			if c.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, c.expectedError)
			}
			if c.expectedCrontab != "" {
				require.EqualValues(t, c.expectedCrontab, actual)
			}
		})
	}
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
