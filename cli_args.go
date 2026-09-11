package main

import (
	"fmt"
	"strconv"
)

type quickArgs struct {
	all     bool
	league  string
	index   *int
	details bool
}

func parseQuickArgs(args []string) (quickArgs, error) {
	var q quickArgs

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			q.all = true
		case "--league":
			i++
			if i >= len(args) {
				return quickArgs{}, fmt.Errorf("--league requires a value")
			}
			q.league = args[i]
		case "--details":
			q.details = true
		default:
			arg := args[i]
			if len(arg) >= 2 && arg[:2] == "--" {
				return quickArgs{}, fmt.Errorf("unknown flag: %s", arg)
			}
			n, err := strconv.Atoi(arg)
			if err != nil {
				return quickArgs{}, fmt.Errorf("invalid argument: %s", arg)
			}
			if q.index != nil {
				return quickArgs{}, fmt.Errorf("unexpected extra argument: %s", arg)
			}
			q.index = &n
		}
	}

	if q.all && (q.league != "" || q.index != nil || q.details) {
		return quickArgs{}, fmt.Errorf("--all cannot be combined with --league, an index, or --details")
	}
	if q.details && q.index == nil {
		return quickArgs{}, fmt.Errorf("--details requires a match index")
	}
	if q.index != nil && q.league == "" {
		return quickArgs{}, fmt.Errorf("a match index requires --league")
	}

	return q, nil
}
