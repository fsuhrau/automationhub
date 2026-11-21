package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/cli/api"
	apiv "github.com/fsuhrau/automationhub/endpoints/api"
	"github.com/spf13/cobra"
)

// runOptions are the flags of the run command.
type runOptions struct {
	binaryPath string
	binaryID   int
	params     string
	tags       string
	async      bool
	timeout    time.Duration
}

// runCommand starts a test and, unless async, waits for its result.
type runCommand struct {
	out       io.Writer
	token     string
	opts      *runOptions
	newStream func(apiURL string) eventStream
}

func newRunCmd(global *globalOptions, opts *runOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run http://localhost:8002 projectID appID testName --binaryID 50 --binary path_to_app --tags \"tag1,tag2,tag3\" --params \"param1=1;parameter2=2\" --async",
		Long:  `Run a new test.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			r := runCommand{out: cmd.OutOrStdout(), token: global.apiToken, opts: opts, newStream: newSSEStream}
			return r.run(cmd.Context(), args)
		},
	}
	f := cmd.PersistentFlags()
	f.StringVar(&opts.binaryPath, "binary", "", "binary /path/to/apk/ipa/zip")
	f.IntVar(&opts.binaryID, "binaryID", 0, "binaryID 123")
	f.StringVar(&opts.params, "params", "", "params \"param1=1;param2=2\"")
	f.StringVar(&opts.tags, "tags", "", "tag \"tag1,tag2,tag3\"")
	f.DurationVar(&opts.timeout, "timeout", 5*time.Minute, "how long to wait for the result")
	return cmd
}

func (r *runCommand) printf(format string, args ...any) {
	fprintf(r.out, format+"\n", args...)
}

func (r *runCommand) run(ctx context.Context, args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("missing parameter")
	}
	t, err := parseTarget(args)
	if err != nil {
		return err
	}
	client := t.client(r.token)
	binaryID, err := r.prepareBinary(ctx, client)
	if err != nil {
		return err
	}
	testID, err := findTestID(ctx, client, args[3])
	if err != nil {
		return err
	}
	return r.execute(ctx, execution{client, t, testID, binaryID})
}

// execution is a test, the binary it runs with and where it runs.
type execution struct {
	client   *api.Client
	target   target
	testID   uint
	binaryID int
}

func (r *runCommand) execute(ctx context.Context, ex execution) error {
	t, ref := ex.target, ex
	r.printf("execute test %d with binaryId: %d\n%s", ref.testID, ref.binaryID, r.opts.params)
	testRun, err := ex.client.ExecuteTest(ctx, ref.testID, apiv.RunTestRequest{AppBinaryID: ref.binaryID, Params: strings.ReplaceAll(r.opts.params, ";", "\n")})
	if err != nil {
		return err
	}
	defer r.printf("Check your restults at: %s/project/%s/app/%d/test/%d/run/%d", t.apiURL, t.projectID, t.appID, ref.testID, testRun.ID)
	if r.opts.async {
		r.printf("Test Running...")
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, r.opts.timeout)
	defer cancel()
	success, err := waiter{stream: r.newStream(t.apiURL), out: r.out}.await(ctx, testRun.ID)
	if err == nil && success {
		r.printf("Test finished Successfully!")
	}
	return err
}

// prepareBinary uploads the binary (and tags it) when one is given and returns
// the id of the binary to run with.
func (r *runCommand) prepareBinary(ctx context.Context, client *api.Client) (int, error) {
	if r.opts.binaryPath == "" {
		return r.opts.binaryID, nil
	}
	r.printf("uploading appBundle")
	bundle, err := client.UploadBundle(ctx, r.opts.binaryPath)
	if err != nil {
		return 0, err
	}
	r.printf("upload finished new binaryID: %d", bundle.ID)
	if r.opts.tags != "" {
		bundle.Tags = r.opts.tags
		if err := client.UpdateApp(ctx, bundle); err != nil {
			return 0, err
		}
	}
	return int(bundle.ID), nil
}

// findTestID is the id of the test called name.
func findTestID(ctx context.Context, client *api.Client, name string) (uint, error) {
	tests, err := client.GetTests(ctx)
	if err != nil {
		return 0, err
	}
	for _, t := range tests {
		if t.Name == name && name != "" {
			return t.ID, nil
		}
	}
	return 0, fmt.Errorf("no test provided or test could not be found")
}
