package cron

import (
	"time"

	robfig "github.com/robfig/cron/v3"
)

type Parser struct {
	parser robfig.Parser
}

func New() *Parser {
	return &Parser{
		parser: robfig.NewParser(robfig.Minute | robfig.Hour | robfig.Dom | robfig.Month | robfig.Dow),
	}
}

func (p *Parser) Validate(expression string) error {
	_, err := p.parser.Parse(expression)
	return err
}

func (p *Parser) Next(expression, timezone string, after time.Time) (time.Time, error) {
	schedule, err := p.parser.Parse(expression)
	if err != nil {
		return time.Time{}, err
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	return schedule.Next(after.In(location)).UTC(), nil
}
