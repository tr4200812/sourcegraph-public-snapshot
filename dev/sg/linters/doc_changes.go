package linters

import (
	"context"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func docChangesLint() *linter {
	return runCheck("Stale doc check", func(ctx context.Context, out *std.Output, state *repo.State) error {
		diff, err := state.GetDiff("doc/**/*.md")
		if err != nil {
			return err
		}
		// If no mardown files were edited, we can exit early
		if len(diff) == 0 {
			return nil
		}
		diffset := make(map[string]struct{}, len(diff))
		for filename := range diff {
			diffset[filename] = struct{}{}
		}

		// List all autogenerated doc files as we will exclude them from this linter
		cmd := []string{"bazel", "cquery", `"filter("\.md", deps(//dev/tools:docsite union
			//doc/admin/observability:doc_files union
        	//doc/cli/references:doc_files union
        	//doc/dev/background-information/telemetry:doc_files))"`, "--output=files"}
		managedDocFiles, err := root.Run(run.Cmd(ctx, cmd...).StdOut()).Lines()
		if err != nil {
			return err
		}

		//remove edited files if they are managed by automation
		for _, managedDoc := range managedDocFiles {
			delete(diffset, managedDoc)
		}

		if len(diffset) > 0 {
			files := make([]string, 0, len(diffset))
			for file := range diffset {
				files = append(files, file)
			}
			return errors.Newf(
				"Your local branch has changes in the doc folder to the listed files:\n%s%s",
				strings.Join(files, "\n"),
				"\n\n`./doc` is deprecated, and new documentation should be commited to https://github.com/sourcegraph/docs")
		}
		return nil
	})
}