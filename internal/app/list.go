package app

import (
	"context"
	"errors"
	"flag"
	"io"
	"net/url"

	"github.com/mattjefferson/oura-cli/internal/oura"
	"github.com/mattjefferson/oura-cli/internal/output"
)

func runList(printer *output.Printer, opts GlobalOptions, args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var startDate string
	var endDate string
	var startDateTime string
	var endDateTime string
	var nextToken string
	var sandbox bool
	var help bool

	fs.StringVar(&startDate, "start-date", "", "start date")
	fs.StringVar(&endDate, "end-date", "", "end date")
	fs.StringVar(&startDateTime, "start-datetime", "", "start datetime")
	fs.StringVar(&endDateTime, "end-datetime", "", "end datetime")
	fs.StringVar(&nextToken, "next-token", "", "next token")
	fs.BoolVar(&sandbox, "sandbox", false, "use sandbox")
	fs.BoolVar(&help, "help", false, "show help")
	fs.BoolVar(&help, "h", false, "show help")

	if err := fs.Parse(args); err != nil {
		printer.Errorf("flag error: %v", err)
		printer.WriteErr("\n")
		printer.WriteErr(listUsage())
		return 2
	}
	if help {
		printer.Write(listUsage())
		return 0
	}

	rest := fs.Args()
	if len(rest) == 0 {
		printer.Errorf("resource required")
		printer.WriteErr("\n")
		printer.WriteErr(listUsage())
		return 2
	}

	resource, ok := oura.LookupResource(rest[0])
	if !ok {
		printer.Errorf("unknown resource: %s", rest[0])
		return 2
	}
	if !resource.SupportsList {
		printer.Errorf("resource is not listable: %s", resource.Key)
		return 2
	}

	query, err := buildListQuery(resource, startDate, endDate, startDateTime, endDateTime, nextToken)
	if err != nil {
		printer.Errorf("invalid query: %v", err)
		return 2
	}

	client, code, err := loadClient(opts, printer)
	if err != nil {
		printer.Errorf("auth required: %v", err)
		return code
	}

	path := oura.BuildPath(sandbox, resource.PathSegment)
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	resp, err := client.Get(ctx, path, query)
	if err != nil {
		printer.Errorf("request failed: %v", err)
		return 4
	}
	if resp.Status >= 400 {
		printer.Errorf("api error (%d): %s", resp.Status, apiErrorMessage(resp.Body))
		return exitCodeForStatus(resp.Status)
	}
	if err := printer.PrintJSON(resp.Body); err != nil {
		printer.Errorf("output failed: %v", err)
		return 1
	}
	return 0
}

func buildListQuery(resource oura.Resource, startDate, endDate, startDateTime, endDateTime, nextToken string) (url.Values, error) {
	params := map[string]string{}
	if nextToken != "" {
		params["next_token"] = nextToken
	}

	switch resource.Query {
	case oura.QueryNone:
		if startDate != "" || endDate != "" || startDateTime != "" || endDateTime != "" {
			return nil, errors.New("resource does not accept date filters")
		}
	case oura.QueryNextTokenOnly:
		if startDate != "" || endDate != "" || startDateTime != "" || endDateTime != "" {
			return nil, errors.New("resource only accepts next_token")
		}
	case oura.QueryDate:
		if startDateTime != "" || endDateTime != "" {
			return nil, errors.New("use --start-date/--end-date for this resource")
		}
		if startDate == "" && endDate == "" {
			break
		}
		if startDate == "" || endDate == "" {
			return nil, errors.New("start-date and end-date must both be set")
		}
		start, err := parseDate(startDate)
		if err != nil {
			return nil, err
		}
		end, err := parseDate(endDate)
		if err != nil {
			return nil, err
		}
		if end.Before(start) {
			return nil, errors.New("end-date must be after start-date")
		}
		params["start_date"] = formatDate(start)
		params["end_date"] = formatDate(end)
	case oura.QueryDateTime:
		if startDate != "" || endDate != "" {
			return nil, errors.New("use --start-datetime/--end-datetime for this resource")
		}
		if startDateTime == "" && endDateTime == "" {
			break
		}
		if startDateTime == "" || endDateTime == "" {
			return nil, errors.New("start-datetime and end-datetime must both be set")
		}
		start, err := parseDateTime(startDateTime)
		if err != nil {
			return nil, err
		}
		end, err := parseDateTime(endDateTime)
		if err != nil {
			return nil, err
		}
		if end.Before(start) {
			return nil, errors.New("end-datetime must be after start-datetime")
		}
		params["start_datetime"] = formatDateTime(start)
		params["end_datetime"] = formatDateTime(end)
	default:
		return nil, errors.New("unknown query type")
	}

	return oura.BuildQuery(params), nil
}
