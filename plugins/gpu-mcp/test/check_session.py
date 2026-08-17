"""Check a gpu-mcp JSON-RPC session transcript (CI smoke test).

test/session.jsonl deliberately contains negative cases (unknown tool, malformed
JSON line), so those error codes are expected. Any other error fails the build.
"""
from __future__ import annotations

import json
import sys

path = sys.argv[1] if len(sys.argv) > 1 else "out.jsonl"

EXPECTED_TOOLS = {
    "gpu_get_status",
    "gpu_query_metrics",
    "gpu_list_processes",
    "gpu_check_requirements",
    "gpu_run_command",
}
# -32602 = unknown tool, -32700 = parse error: both are exercised on purpose
ALLOWED_ERROR_CODES = {-32602, -32700}

with open(path, encoding="utf-8") as fh:
    messages = [json.loads(line) for line in fh if line.strip()]

tools: list[str] = []
results = 0
for msg in messages:
    result = msg.get("result")
    if isinstance(result, dict):
        results += 1
        if "tools" in result:
            tools = [tool["name"] for tool in result["tools"]]

errors = [msg["error"] for msg in messages if "error" in msg]
unexpected = [err for err in errors if err.get("code") not in ALLOWED_ERROR_CODES]
missing = EXPECTED_TOOLS - set(tools)

print(f"responses: {len(messages)} (results: {results}, errors: {len(errors)})")
print(f"tools: {sorted(tools)}")
for err in errors:
    print(f"expected negative case: code {err.get('code')} — {err.get('message')}")

if not messages:
    sys.exit("no JSON-RPC responses at all")
if missing:
    sys.exit(f"missing tools: {sorted(missing)}")
if results < 5:
    sys.exit(f"expected at least 5 successful results, got {results}")
if unexpected:
    sys.exit(f"unexpected JSON-RPC errors: {unexpected}")
if not errors:
    sys.exit("negative test cases produced no errors — session file changed?")

print("smoke test ok")