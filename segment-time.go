package main

import (
	"time"

	pwl "github.com/justjanne/powerline-go/powerline"
)

func segmentTime(p *powerline) {
	p.appendSegment("time", pwl.Segment{
		Content:    time.Now().Format("02T15:04:05"),
		Foreground: p.theme.TimeFg,
		Background: p.theme.TimeBg,
	})
}
