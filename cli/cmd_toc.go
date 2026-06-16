package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// tocCmd is an alias for chapters.
func (a *App) tocCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toc",
		Short: "List all chapters (table of contents)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			chapters, err := a.client.Chapters(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			if a.limit > 0 && len(chapters) > a.limit {
				chapters = chapters[:a.limit]
			}
			return a.renderOrEmpty(chapters, len(chapters))
		},
	}
	return cmd
}

func (a *App) chapterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chapter <num>",
		Short: "Show info for a chapter by number",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				return codeError(exitUsage, fmt.Errorf("chapter number must be an integer: %w", err))
			}
			ch, err := a.client.ChapterByNum(cmd.Context(), num)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render([]*chapStub{{
				Rank:    ch.Rank,
				Chapter: ch.Chapter,
				Part:    ch.Part,
				Title:   ch.Title,
				PDF:     ch.PDF,
			}})
		},
	}
	return cmd
}

// chapStub is a render-friendly wrapper (same fields as Chapter).
type chapStub struct {
	Rank    int    `json:"rank"    csv:"rank"    tsv:"rank"`
	Chapter int    `json:"chapter" csv:"chapter" tsv:"chapter"`
	Part    string `json:"part"    csv:"part"    tsv:"part"`
	Title   string `json:"title"   csv:"title"   tsv:"title"`
	PDF     string `json:"pdf"     csv:"pdf"     tsv:"pdf"`
}

func (a *App) searchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search chapters by title or part name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := a.client.Search(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			if a.limit > 0 && len(results) > a.limit {
				results = results[:a.limit]
			}
			return a.renderOrEmpty(results, len(results))
		},
	}
	return cmd
}

func (a *App) infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print site stats (chapter count, parts)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := a.client.Info(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(info)
		},
	}
}
