package docs_validation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractPropagateBlocks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantCount  int
		wantFirst  *PropagateBlock // optional check for first block.
		wantSecond *PropagateBlock // optional check for second block.
	}{
		{
			name: "basic single block",
			content: join(
				"# Heading",
				`<!-- @propagate to="file.md" as="chunk-1" -->`,
				"Line one",
				"Line two",
				"<!-- @/propagate -->",
				"Other text",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{TargetFile: "file.md", ChunkID: "chunk-1", Content: "Line one\nLine two\n", LineNumber: 2},
		},
		{
			name: "multiple blocks",
			content: join(
				`<!-- @propagate to="a.md" as="alpha" -->`,
				"content-a",
				"<!-- @/propagate -->",
				"gap",
				`<!-- @propagate to="b.md" as="beta" -->`,
				"content-b",
				"<!-- @/propagate -->",
			),
			wantCount:  2,
			wantFirst:  &PropagateBlock{TargetFile: "a.md", ChunkID: "alpha", Content: "content-a\n"},
			wantSecond: &PropagateBlock{TargetFile: "b.md", ChunkID: "beta", Content: "content-b\n"},
		},
		{
			name: "skips markers inside code fences",
			content: join(
				"```yaml",
				`<!-- @propagate to="skipped.md" as="skipped" -->`,
				"should not match",
				"<!-- @/propagate -->",
				"```",
				`<!-- @propagate to="real.md" as="real" -->`,
				"real content",
				"<!-- @/propagate -->",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{TargetFile: "real.md", ChunkID: "real", Content: "real content\n"},
		},
		{
			name: "preserves code fences inside propagated content",
			content: join(
				`<!-- @propagate to="target.md" as="with-fence" -->`,
				"**Example**:",
				"",
				"```bash",
				`echo "hello"`,
				"```",
				"",
				"Done.",
				"<!-- @/propagate -->",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{
				TargetFile: "target.md",
				ChunkID:    "with-fence",
				Content:    "**Example**:\n\n```bash\necho \"hello\"\n```\n\nDone.\n",
			},
		},
		{
			name:      "no markers",
			content:   "no markers here\njust text",
			wantCount: 0,
		},
		{
			name: "empty content block",
			content: join(
				`<!-- @propagate to="f.md" as="empty" -->`,
				"<!-- @/propagate -->",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{TargetFile: "f.md", ChunkID: "empty", Content: "\n"},
		},
		{
			name: "multi-target comma separated",
			content: join(
				`<!-- @propagate to="a.md, b.md" as="shared" -->`,
				"Shared content",
				"<!-- @/propagate -->",
			),
			wantCount:  2,
			wantFirst:  &PropagateBlock{TargetFile: "a.md", ChunkID: "shared", Content: "Shared content\n", LineNumber: 1},
			wantSecond: &PropagateBlock{TargetFile: "b.md", ChunkID: "shared", Content: "Shared content\n", LineNumber: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			blocks := extractPropagateBlocks(tc.content)
			require.Len(t, blocks, tc.wantCount)

			if tc.wantFirst != nil && tc.wantCount >= 1 {
				require.Equal(t, tc.wantFirst.TargetFile, blocks[0].TargetFile)
				require.Equal(t, tc.wantFirst.ChunkID, blocks[0].ChunkID)
				require.Equal(t, tc.wantFirst.Content, blocks[0].Content)

				if tc.wantFirst.LineNumber > 0 {
					require.Equal(t, tc.wantFirst.LineNumber, blocks[0].LineNumber)
				}
			}

			if tc.wantSecond != nil && tc.wantCount >= 2 {
				require.Equal(t, tc.wantSecond.TargetFile, blocks[1].TargetFile)
				require.Equal(t, tc.wantSecond.ChunkID, blocks[1].ChunkID)
				require.Equal(t, tc.wantSecond.Content, blocks[1].Content)
			}
		})
	}
}

func TestExtractSourceBlocks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		wantCount  int
		wantFirst  *SourceBlock
		wantSecond *SourceBlock
	}{
		{
			name: "basic single block",
			content: join(
				"Some intro",
				`<!-- @source from="docs/ENG-HANDBOOK.md" as="chunk-1" -->`,
				"Source line one",
				"Source line two",
				"<!-- @/source -->",
				"See more",
			),
			wantCount: 1,
			wantFirst: &SourceBlock{ChunkID: "chunk-1", Content: "Source line one\nSource line two\n", LineNumber: 2},
		},
		{
			name: "multiple blocks",
			content: join(
				`<!-- @source from="arch.md" as="alpha" -->`,
				"a-content",
				"<!-- @/source -->",
				"glue text",
				`<!-- @source from="arch.md" as="beta" -->`,
				"b-content",
				"<!-- @/source -->",
			),
			wantCount:  2,
			wantFirst:  &SourceBlock{ChunkID: "alpha", Content: "a-content\n"},
			wantSecond: &SourceBlock{ChunkID: "beta", Content: "b-content\n"},
		},
		{
			name:      "no markers",
			content:   "no markers here",
			wantCount: 0,
		},
		{
			name: "empty content block",
			content: join(
				`<!-- @source from="arch.md" as="empty" -->`,
				"<!-- @/source -->",
			),
			wantCount: 1,
			wantFirst: &SourceBlock{ChunkID: "empty", Content: "\n"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			blocks := extractSourceBlocks(tc.content)
			require.Len(t, blocks, tc.wantCount)

			if tc.wantFirst != nil && tc.wantCount >= 1 {
				require.Equal(t, tc.wantFirst.ChunkID, blocks[0].ChunkID)
				require.Equal(t, tc.wantFirst.Content, blocks[0].Content)

				if tc.wantFirst.LineNumber > 0 {
					require.Equal(t, tc.wantFirst.LineNumber, blocks[0].LineNumber)
				}
			}

			if tc.wantSecond != nil && tc.wantCount >= 2 {
				require.Equal(t, tc.wantSecond.ChunkID, blocks[1].ChunkID)
				require.Equal(t, tc.wantSecond.Content, blocks[1].Content)
			}
		})
	}
}
