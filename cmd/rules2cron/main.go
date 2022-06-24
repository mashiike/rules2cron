package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fujiwara/logutils"
	"github.com/mashiike/rules2cron"
)

var (
	Version string = "current"
)

func main() {
	var (
		refDate      string
		tz           string
		minLevel     string
		showDisabled bool
	)
	flag.CommandLine.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "rules2json is cron-like notation converter for ScheduleExpression in EventBridge's Rule")
		fmt.Fprintln(flag.CommandLine.Output(), "version:", Version)
		flag.CommandLine.PrintDefaults()
	}
	flag.StringVar(&minLevel, "log-level", "info", "rules2json log level")
	flag.StringVar(&refDate, "ref-date", time.Now().Format("2006-01-01"), "date of conversion basis")
	flag.StringVar(&tz, "tz", "UTC", "Which time zone to convert to")
	flag.BoolVar(&showDisabled, "show-disabled", false, "show disabled rules")
	flag.Parse()

	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{"debug", "info", "notice", "warn", "error"},
		ModifierFuncs: []logutils.ModifierFunc{
			logutils.Color(color.FgHiBlack),
			nil,
			logutils.Color(color.FgHiBlue),
			logutils.Color(color.FgYellow),
			logutils.Color(color.FgRed, color.BgBlack),
		},
		MinLevel: logutils.LogLevel(strings.ToLower(minLevel)),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	date, err := time.Parse("2006-01-02", refDate)
	if err != nil {
		log.Fatalln("[error] ", err)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Println("[wan] can not load location, use UTC: ", err)
		loc = time.UTC
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	app, err := rules2cron.New(ctx, &rules2cron.Converter{
		ReferenceDate: date,
		TimeZone:      loc,
	})
	if err != nil {
		log.Fatalln("[error] ", err)
	}
	if err := app.RunWithContext(ctx, os.Stdout, showDisabled); err != nil {
		log.Fatalln("[error] ", err)
	}
}
