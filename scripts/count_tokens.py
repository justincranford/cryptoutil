"""
Count tokens for files or glob of files using tiktoken and estimate chat message token costs.

Usage examples (PowerShell):

# Create venv and install:
python -m venv .venv; . \.venv\Scripts\Activate.ps1; python -m pip install -U pip setuptools wheel; python -m pip install tiktoken

# Count tokens for all instruction files as system messages for gpt-4o and show a Claude Sonnet 4 example cost
python .\scripts\count_tokens.py --model gpt-4o --glob ".github/instructions/*.md" --as-message system --price-per-1k 0.01

# Count tokens for a single file as plain text (no message framing)
python .\scripts\count_tokens.py --file .github/copilot-instructions.md --as-message none --model gpt-4o

Notes:
- The default --price-per-1k is 0.01 USD in this script as an example for Claude Sonnet 4. Replace with the actual Claude Sonnet 4 price per 1k tokens if you have it.

Copilot pricing references & findings:
- GitHub Copilot product page: https://github.com/features/copilot
- GitHub Copilot plans comparison: https://github.com/features/copilot/plans
- GitHub pricing landing: https://github.com/pricing

Findings (as of Oct 6, 2025):
- Copilot Free includes 50 agent/chat requests per month and 2,000 completions per month. (See the Copilot product page.)
- Copilot Pro ($10/month) advertises "6x more premium requests than Copilot Free". Interpreting that multiplier against the Free plan's 50 chat requests gives 300 premium requests/month for Pro. Pro also advertises "Unlimited completions and chats" for general usage and provides access to Claude Sonnet 4, GPT-5, and other premium models.
- Source lines used to reach this conclusion are present on the Copilot product and pricing pages linked above. If you need an official billing/FAQ confirmation, GitHub support or the Copilot Trust Center is recommended.
"""
from __future__ import annotations
import argparse
import glob
import os
from pathlib import Path
from typing import List, Dict

try:
    import tiktoken
except Exception:
    tiktoken = None


def ensure_tiktoken():
    if tiktoken is None:
        raise RuntimeError("tiktoken is not installed. Run: pip install tiktoken")


# Based on tiktoken/OpenAI guidance for chat message token counting
def num_tokens_from_messages(messages: List[Dict[str, str]], model: str = "gpt-3.5-turbo-0301") -> int:
    ensure_tiktoken()
    try:
        encoding = tiktoken.encoding_for_model(model)
    except Exception:
        encoding = tiktoken.get_encoding("cl100k_base")

    # Defaults taken from OpenAI guidance for gpt-3.5 and gpt-4 families.
    if model.startswith("gpt-3.5") or model.startswith("gpt-4") or model.startswith("gpt-4o"):
        tokens_per_message = 3
        tokens_per_name = 1
    else:
        tokens_per_message = 3
        tokens_per_name = 1

    num_tokens = 0
    for message in messages:
        num_tokens += tokens_per_message
        for key, value in message.items():
            # value is a string
            num_tokens += len(encoding.encode(value))
            if key == "name":
                num_tokens += tokens_per_name
    num_tokens += 3  # every reply is primed with <|assistant|>
    return num_tokens


def num_tokens_from_text(text: str, model: str = "gpt-3.5-turbo-0301") -> int:
    ensure_tiktoken()
    try:
        encoding = tiktoken.encoding_for_model(model)
    except Exception:
        encoding = tiktoken.get_encoding("cl100k_base")
    return len(encoding.encode(text))


def analyze_files(paths: List[Path], model: str, as_message: str):
    rows = []
    total_tokens = 0
    total_chars = 0
    for p in paths:
        try:
            # read with replacement for invalid bytes to avoid crashes on binary/BOM files
            text = p.read_text(encoding="utf-8", errors="replace")
        except TypeError:
            # older Python versions of Path.read_text may not accept errors kwarg
            with p.open('rb') as fh:
                raw = fh.read()
            text = raw.decode('utf-8', errors='replace')
        # detect if replacement characters were introduced (U+FFFD)
        if '\ufffd' in text:
            # keep going but warn
            print(f"Warning: file {p} contained invalid UTF-8 bytes and was read with replacements.")
        chars = len(text)
        total_chars += chars
        if as_message == "system":
            messages = [{"role": "system", "content": text}]
            tokens = num_tokens_from_messages(messages, model=model)
        elif as_message == "user":
            messages = [{"role": "user", "content": text}]
            tokens = num_tokens_from_messages(messages, model=model)
        else:
            tokens = num_tokens_from_text(text, model=model)
        rows.append({"path": str(p), "chars": chars, "tokens": tokens})
        total_tokens += tokens
    return rows, total_tokens, total_chars


def find_paths(file: str | None, glob_pattern: str | None) -> List[Path]:
    paths: List[Path] = []
    if file:
        p = Path(file)
        if p.exists():
            paths.append(p)
        else:
            raise FileNotFoundError(f"File not found: {file}")
    if glob_pattern:
        matches = glob.glob(glob_pattern, recursive=True)
        for m in matches:
            paths.append(Path(m))
    # remove duplicates and sort
    unique = []
    seen = set()
    for p in paths:
        rp = str(p.resolve())
        if rp not in seen:
            seen.add(rp)
            unique.append(p)
    return unique


def main():
    parser = argparse.ArgumentParser(description="Count tokens for files using tiktoken")
    parser.add_argument("--model", default="gpt-4o", help="Model name to pick encoding rules (default: gpt-4o)")
    parser.add_argument("--file", help="Single file to analyze")
    parser.add_argument("--glob", help="Glob pattern to include (e.g. '.github/instructions/*.md')")
    parser.add_argument("--as-message", choices=["system", "user", "none"], default="system",
                        help="Treat file content as a chat message of given role (adds message framing tokens). 'none' counts raw tokens")
    parser.add_argument("--price-per-1k", type=float, default=0.01, help="Price in USD per 1000 tokens (default: 0.01 example)")
    parser.add_argument("--show-raw", action="store_true", help="Only print raw per-file token counts without table formatting")
    parser.add_argument("--show-sum-only", action="store_true", help="Only print totals")
    args = parser.parse_args()

    try:
        paths = find_paths(args.file, args.glob)
    except FileNotFoundError as e:
        print(e)
        return

    if not paths:
        print("No files found. Provide --file or --glob that matches files.")
        return

    if tiktoken is None:
        print("tiktoken not installed. Install with: pip install tiktoken")
        return

    rows, total_tokens, total_chars = analyze_files(paths, model=args.model, as_message=args.as_message)

    if not args.show_sum_only:
        if args.show_raw:
            print(f"Per-file token counts (model={args.model}, as_message={args.as_message}):")
            for r in rows:
                cpt = r["chars"] / r["tokens"] if r["tokens"] > 0 else float('inf')
                print(f"- {r['path']}: {r['tokens']} tokens, {r['chars']} chars ({cpt:.1f} chars/token)")
            print("")
        else:
            # Nicely formatted table with cost estimates
            print(f"Per-file token counts and cost estimates (model={args.model}, as_message={args.as_message})")
            print(f"{'Path':<80} {'Tokens':>8} {'Chars':>10} {'Chars/Token':>12} {'Cost(USD)':>10}")
            print('-' * 120)
            for r in rows:
                cpt = r["chars"] / r["tokens"] if r["tokens"] > 0 else float('inf')
                cost = (r["tokens"] / 1000.0) * args.price_per_1k
                print(f"{r['path']:<80} {r['tokens']:8d} {r['chars']:10d} {cpt:12.2f} {cost:10.4f}")
            print('')

    avg_chars_per_token = total_chars / total_tokens if total_tokens > 0 else float('inf')
    print(f"Total files: {len(rows)}")
    print(f"Total tokens: {total_tokens}")
    print(f"Total chars: {total_chars}")
    print(f"Average chars/token across selection: {avg_chars_per_token:.2f}")
    total_cost = (total_tokens / 1000.0) * args.price_per_1k
    print(f"Estimated cost (@ {args.price_per_1k} USD per 1k tokens): {total_cost:.5f} USD")

    # quick heuristic guidance
    print("\nHeuristics:")
    print("- Roughly 3-4 chars per token for English prose; your data may differ.")
    print("- If using chat-based API, choose --as-message to include framing tokens per message (~3 tokens/msg + name handling).")
    print("- To estimate cost multiply token count by model price per 1k tokens (consult your model pricing).")


if __name__ == '__main__':
    main()
