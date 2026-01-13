package app

import (
	"context"
	"flag"
	"io"

	"github.com/mattjefferson/oura-cli/internal/oura"
	"github.com/mattjefferson/oura-cli/internal/output"
)

func runGet(printer *output.Printer, opts GlobalOptions, args []string) int {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var sandbox bool
	var help bool

	fs.BoolVar(&sandbox, "sandbox", false, "use sandbox")
	fs.BoolVar(&help, "help", false, "show help")
	fs.BoolVar(&help, "h", false, "show help")

	if err := fs.Parse(args); err != nil {
		printer.Errorf("flag error: %v", err)
		printer.WriteErr("\n")
		printer.WriteErr(getUsage())
		return 2
	}
	if help {
		printer.Write(getUsage())
		return 0
	}

	rest := fs.Args()
	if len(rest) == 0 {
		printer.Errorf("resource required")
		printer.WriteErr("\n")
		printer.WriteErr(getUsage())
		return 2
	}

	resource, ok := oura.LookupResource(rest[0])
	if !ok {
		printer.Errorf("unknown resource: %s", rest[0])
		return 2
	}
	if !resource.SupportsGet {
		printer.Errorf("resource is not fetchable: %s", resource.Key)
		return 2
	}

	documentID := ""
	if resource.Key != "personal_info" {
		if len(rest) < 2 {
			printer.Errorf("document_id required for %s", resource.Key)
			return 2
		}
		documentID = rest[1]
	} else if len(rest) > 1 {
		printer.Errorf("personal_info does not accept a document_id")
		return 2
	}

	return fetchResource(printer, opts, resource, documentID, sandbox)
}

func runWhoami(printer *output.Printer, opts GlobalOptions) int {
	resource, _ := oura.LookupResource("personal_info")
	return fetchResource(printer, opts, resource, "", false)
}

func fetchResource(printer *output.Printer, opts GlobalOptions, resource oura.Resource, documentID string, sandbox bool) int {
	client, code, err := loadClient(opts, printer)
	if err != nil {
		printer.Errorf("auth required: %v", err)
		return code
	}

	path := ""
	if documentID == "" && resource.Key == "personal_info" {
		path = oura.BuildPath(sandbox, resource.PathSegment)
	} else {
		path = oura.BuildDocumentPath(sandbox, resource.PathSegment, documentID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	resp, err := client.Get(ctx, path, nil)
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
