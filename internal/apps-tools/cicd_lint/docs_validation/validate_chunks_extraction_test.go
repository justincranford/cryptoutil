package docs_validation

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testNewlineContent = "\n"

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
			name: testBasicSingleBlock,
			content: join(
				"# Heading",
				`<!-- @to-appendix as="chunk-1" appendixes="file.md" -->`,
				"Line one",
				"Line two",
				"<!-- @/to-appendix -->",
				"Other text",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{TargetFile: testFileMd, ChunkID: testChunk1, Content: "Line one\nLine two\n", LineNumber: 2},
		},
		{
			name: testMultipleBlocks,
			content: join(
				`<!-- @to-appendix as="alpha" appendixes="a.md" -->`,
				"content-a",
				"<!-- @/to-appendix -->",
				"gap",
				`<!-- @to-appendix as="beta" appendixes="b.md" -->`,
				"content-b",
				"<!-- @/to-appendix -->",
			),
			wantCount:  2,
			wantFirst:  &PropagateBlock{TargetFile: testFileA, ChunkID: testAlpha, Content: "content-a\n"},
			wantSecond: &PropagateBlock{TargetFile: testFileB, ChunkID: testBeta, Content: "content-b\n"},
		},
		{
			name: "skips markers inside code fences",
			content: join(
				"```yaml",
				`<!-- @to-appendix as="skipped" appendixes="skipped.md" -->`,
				"should not match",
				"<!-- @/to-appendix -->",
				"```",
				`<!-- @to-appendix as="real" appendixes="real.md" -->`,
				"real content",
				"<!-- @/to-appendix -->",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{TargetFile: "real.md", ChunkID: "real", Content: "real content\n"},
		},
		{
			name: "preserves code fences inside propagated content",
			content: join(
				`<!-- @to-appendix as="with-fence" appendixes="target.md" -->`,
				"**Example**:",
				"",
				"```bash",
				`echo "hello"`,
				"```",
				"",
				"Done.",
				"<!-- @/to-appendix -->",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{
				TargetFile: testTargetMd,
				ChunkID:    "with-fence",
				Content:    "**Example**:\n\n```bash\necho \"hello\"\n```\n\nDone.\n",
			},
		},
		{
			name:      testNoMarkers,
			content:   "no markers here\njust text",
			wantCount: 0,
		},
		{
			name: testEmptyContentBlock,
			content: join(
				`<!-- @to-appendix as="empty" appendixes="f.md" -->`,
				"<!-- @/to-appendix -->",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{TargetFile: "f.md", ChunkID: cryptoutilSharedMagic.EMPTY, Content: testNewlineContent},
		},
		{
			name: "multi-target comma separated",
			content: join(
				`<!-- @to-appendix as="shared" appendixes="a.md, b.md" -->`,
				"Shared content",
				"<!-- @/to-appendix -->",
			),
			wantCount:  2,
			wantFirst:  &PropagateBlock{TargetFile: testFileA, ChunkID: testShared, Content: testSharedContent, LineNumber: 1},
			wantSecond: &PropagateBlock{TargetFile: testFileB, ChunkID: testShared, Content: testSharedContent, LineNumber: 1},
		},
		{
			name: "to-appendix marker",
			content: join(
				`<!-- @to-appendix as="rfc-2119-keywords" appendixes="target.md" -->`,
				"Terminology content",
				"<!-- @/to-appendix -->",
			),
			wantCount: 1,
			wantFirst: &PropagateBlock{
				TargetFile: testTargetMd,
				ChunkID:    testRfc2119Keywords,
				Content:    "Terminology content\n",
				LineNumber: 1,
			},
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
			name: testBasicSingleBlock,
			content: join(
				"Some intro",
				`<!-- @from-eng-handbook as="chunk-1" -->`,
				"Source line one",
				"Source line two",
				"<!-- @/from-eng-handbook -->",
				"See more",
			),
			wantCount: 1,
			wantFirst: &SourceBlock{ChunkID: testChunk1, Content: "Source line one\nSource line two\n", LineNumber: 2},
		},
		{
			name: testMultipleBlocks,
			content: join(
				`<!-- @from-eng-handbook as="alpha" -->`,
				"a-content",
				"<!-- @/from-eng-handbook -->",
				"glue text",
				`<!-- @from-eng-handbook as="beta" -->`,
				"b-content",
				"<!-- @/from-eng-handbook -->",
			),
			wantCount:  2,
			wantFirst:  &SourceBlock{ChunkID: testAlpha, Content: "a-content\n"},
			wantSecond: &SourceBlock{ChunkID: testBeta, Content: "b-content\n"},
		},
		{
			name:      testNoMarkers,
			content:   "no markers here",
			wantCount: 0,
		},
		{
			name: testEmptyContentBlock,
			content: join(
				`<!-- @from-eng-handbook as="empty" -->`,
				"<!-- @/from-eng-handbook -->",
			),
			wantCount: 1,
			wantFirst: &SourceBlock{ChunkID: cryptoutilSharedMagic.EMPTY, Content: testNewlineContent},
		},
		{
			name: "only handbook-derived-body is strict scope",
			content: join(
				"<!-- @from-eng-handbook as=\"outside\" -->",
				"Outside",
				"<!-- @/from-eng-handbook -->",
				"<!-- @handbook-derived-body:start -->",
				"<!-- @from-eng-handbook as=\"inside\" -->",
				"Inside",
				"<!-- @/from-eng-handbook -->",
				"<!-- @handbook-derived-body:end -->",
			),
			wantCount: 1,
			wantFirst: &SourceBlock{ChunkID: "inside", Content: "Inside\n", LineNumber: cryptoutilSharedMagic.IdentityDefaultMaxIdleConns},
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
