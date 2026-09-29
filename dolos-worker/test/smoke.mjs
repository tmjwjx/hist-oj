import assert from "node:assert/strict";
import { spawn } from "node:child_process";

const port = 18_091;
const child = spawn(process.execPath, ["server.mjs"], {
  cwd: new URL("..", import.meta.url),
  env: { ...process.env, HOST: "127.0.0.1", PORT: String(port), MAX_CONCURRENT_ANALYSES: "1" },
  stdio: ["ignore", "pipe", "pipe"],
});

let output = "";
child.stdout.on("data", (chunk) => { output += chunk.toString(); });
child.stderr.on("data", (chunk) => { output += chunk.toString(); });

try {
  await waitForHealth(`http://127.0.0.1:${port}/healthz`);
  const health = await request("GET", `http://127.0.0.1:${port}/healthz`);
  assert.equal(health.status, 200);
  assert.equal(health.body.ok, true);
  assert.equal(health.body.engine, "@dodona/dolos-lib@3.5.1");

  const result = await request("POST", `http://127.0.0.1:${port}/v1/analyze`, {
    language: "Python3",
    submissions: [
      { id: "py-1", filename: "first.py", code: "def add(a, b):\n    return a + b\nprint(add(1, 2))\n" },
      { id: "py-2", filename: "second.py", code: "def sum_values(left, right):\n    return left + right\nprint(sum_values(1, 2))\n" },
    ],
  });
  assert.equal(result.status, 200, JSON.stringify(result.body));
  assert.equal(result.body.language, "python");
  assert.equal(result.body.pairs.length, 1);
  assert.equal(result.body.pairs[0].leftSubmissionId, "py-1");
  assert.equal(result.body.pairs[0].rightSubmissionId, "py-2");
  assert.ok(result.body.pairs[0].similarity > 0);
  assert.equal(result.body.pairs[0].similarity1to2, 100);
  assert.equal(result.body.pairs[0].similarity2to1, 100);
  assert.ok(result.body.pairs[0].fragments.length > 0);

  const cppResult = await request("POST", `http://127.0.0.1:${port}/v1/analyze`, {
    language: "C++ With O2",
    submissions: [
      { id: "cpp-1", filename: "first.cpp", code: "int add(int a, int b) { return a + b; }" },
      { id: "cpp-2", filename: "second.cpp", code: "int sum(int left, int right) { return left + right; }" },
    ],
  });
  assert.equal(cppResult.status, 200, JSON.stringify(cppResult.body));
  assert.equal(cppResult.body.language, "cpp");

  const invalid = await request("POST", `http://127.0.0.1:${port}/v1/analyze`, {
    language: "python",
    submissions: [{ id: "only", filename: "only.py", code: "print(1)" }],
  });
  assert.equal(invalid.status, 400);
  assert.equal(invalid.body.error.code, "invalid_request");

  const unsupported = await request("POST", `http://127.0.0.1:${port}/v1/analyze`, {
    language: "not-a-real-language",
    submissions: [
      { id: "bad-1", filename: "first.txt", code: "one" },
      { id: "bad-2", filename: "second.txt", code: "two" },
    ],
  });
  assert.equal(unsupported.status, 422);
  assert.equal(unsupported.body.error.code, "analysis_failed");

  process.stdout.write("Dolos worker smoke test passed\n");
} finally {
  child.kill("SIGTERM");
  await new Promise((resolve) => child.once("exit", resolve));
  if (output.includes("UnhandledPromiseRejection")) {
    process.stderr.write(output);
    process.exitCode = 1;
  }
}

async function waitForHealth(url) {
  const deadline = Date.now() + 30_000;
  while (Date.now() < deadline) {
    try {
      const response = await fetch(url);
      if (response.ok) return;
    } catch {
      // The worker is still starting and loading native parsers.
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`worker did not start in time: ${output}`);
}

async function request(method, url, body) {
  const response = await fetch(url, {
    method,
    headers: body ? { "content-type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  return { status: response.status, body: await response.json() };
}
