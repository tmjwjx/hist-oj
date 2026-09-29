import http from "node:http";
import { Dolos, File } from "@dodona/dolos-lib";

const SERVICE_NAME = "hist-oj-dolos-worker";
const SERVICE_VERSION = "0.1.0";
const ENGINE_PACKAGE = "@dodona/dolos-lib";
const ENGINE_VERSION = "3.5.1";

const HOST = process.env.HOST || "127.0.0.1";
const PORT = parsePort(process.env.PORT, 8091);
const MAX_BODY_BYTES = parsePositiveInteger(process.env.MAX_BODY_BYTES, 128 * 1024 * 1024);
const MAX_SUBMISSIONS = parsePositiveInteger(process.env.MAX_SUBMISSIONS, 1_000);
const MAX_CODE_BYTES = parsePositiveInteger(process.env.MAX_CODE_BYTES, 64 * 1024 * 1024);
const MAX_PAIRS = parsePositiveInteger(process.env.MAX_PAIRS, 200_000);
const MAX_FRAGMENTS = parsePositiveInteger(process.env.MAX_FRAGMENTS, 200);
const MAX_CONCURRENT_ANALYSES = parsePositiveInteger(process.env.MAX_CONCURRENT_ANALYSES, 1);
const ANALYSIS_TIMEOUT_MS = parsePositiveInteger(process.env.ANALYSIS_TIMEOUT_MS, 900_000);
const AUTH_TOKEN = typeof process.env.AUTH_TOKEN === "string" ? process.env.AUTH_TOKEN.trim() : "";

let activeAnalyses = 0;

const LANGUAGE_ALIASES = new Map([
  ["c++", "cpp"],
  ["c++11", "cpp"],
  ["c++14", "cpp"],
  ["c++17", "cpp"],
  ["c++20", "cpp"],
  ["c++23", "cpp"],
  ["cpp11", "cpp"],
  ["cpp14", "cpp"],
  ["cpp17", "cpp"],
  ["cpp20", "cpp"],
  ["cpp23", "cpp"],
  ["csharp", "c_sharp"],
  ["c#", "c_sharp"],
  ["golang", "go"],
  ["python2", "python"],
  ["python3", "python"],
  ["pypy", "python"],
  ["pypy3", "python"],
  ["javascriptnode", "javascript"],
  ["javascriptv8", "javascript"],
  ["nodejs", "javascript"],
  ["typescriptnode", "typescript"],
]);

const LANGUAGE_EXTENSIONS = new Map([
  ["bash", ".sh"],
  ["c", ".c"],
  ["cpp", ".cpp"],
  ["c_sharp", ".cs"],
  ["elm", ".elm"],
  ["go", ".go"],
  ["groovy", ".groovy"],
  ["java", ".java"],
  ["javascript", ".js"],
  ["modelica", ".mo"],
  ["ocaml", ".ml"],
  ["php", ".php"],
  ["python", ".py"],
  ["r", ".r"],
  ["rust", ".rs"],
  ["scala", ".scala"],
  ["sql", ".sql"],
  ["typescript", ".ts"],
  ["verilog", ".v"],
]);

function parsePositiveInteger(value, fallback) {
  if (value === undefined || value === "") {
    return fallback;
  }
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function parsePort(value, fallback) {
  const parsed = parsePositiveInteger(value, fallback);
  return parsed <= 65_535 ? parsed : fallback;
}

function normalizeLanguage(value) {
  if (value === undefined || value === null || value === "") {
    return null;
  }
  if (typeof value !== "string") {
    throw clientError("language must be a string");
  }
  const language = value.trim().toLowerCase();
  if (!/^[a-z0-9_+#.-]{1,64}$/.test(language.replaceAll(" ", ""))) {
    throw clientError("language contains unsupported characters");
  }
  const compact = language.replace(/[\s-]+/g, "");
  const alias = LANGUAGE_ALIASES.get(compact) || LANGUAGE_ALIASES.get(language);
  if (alias) {
    return alias;
  }
  // Judge environments often append compiler/runtime labels (for example
  // "C++ With O2" or "GNU C++17"). Keep those labels out of Dolos' parser
  // lookup while still rejecting arbitrary values above.
  if (compact.startsWith("c++") || compact.startsWith("cpp") || compact.includes("gnuc++")) {
    return "cpp";
  }
  if (compact === "c" || compact.startsWith("cwith")) {
    return "c";
  }
  if (compact.startsWith("python") || compact.startsWith("pypy")) {
    return "python";
  }
  if (compact.startsWith("javascript") || compact.startsWith("nodejs")) {
    return "javascript";
  }
  if (compact.startsWith("typescript")) {
    return "typescript";
  }
  if (compact.startsWith("golang") || compact === "go") {
    return "go";
  }
  if (compact.startsWith("csharp") || compact.startsWith("c#")) {
    return "c_sharp";
  }
  return language;
}

function clientError(message) {
  const error = new Error(message);
  error.statusCode = 400;
  error.code = "invalid_request";
  return error;
}

function safeInteger(value, name, { min = 0, max = Number.MAX_SAFE_INTEGER } = {}) {
  if (value === undefined || value === null) {
    return undefined;
  }
  if (!Number.isSafeInteger(value) || value < min || value > max) {
    throw clientError(`${name} must be an integer between ${min} and ${max}`);
  }
  return value;
}

function safeFraction(value, name) {
  if (value === undefined || value === null) {
    return undefined;
  }
  if (typeof value !== "number" || !Number.isFinite(value) || value < 0 || value > 1) {
    throw clientError(`${name} must be a number between 0 and 1`);
  }
  return value;
}

function normalizeOptions(input) {
  if (input === undefined) {
    return {};
  }
  if (!input || typeof input !== "object" || Array.isArray(input)) {
    throw clientError("options must be an object");
  }
  const options = {};
  const kgramLength = safeInteger(input.kgramLength, "options.kgramLength", { min: 1, max: 256 });
  const kgramsInWindow = safeInteger(input.kgramsInWindow, "options.kgramsInWindow", { min: 1, max: 512 });
  const limitResults = safeInteger(input.limitResults, "options.limitResults", { min: 1, max: MAX_PAIRS });
  const maxFingerprintCount = safeInteger(input.maxFingerprintCount, "options.maxFingerprintCount", { min: 1, max: MAX_SUBMISSIONS });
  const minFragmentLength = safeInteger(input.minFragmentLength, "options.minFragmentLength", { min: 1, max: 10_000 });
  const minSimilarity = safeFraction(input.minSimilarity, "options.minSimilarity");
  const maxFingerprintPercentage = safeFraction(input.maxFingerprintPercentage, "options.maxFingerprintPercentage");

  if (kgramLength !== undefined) options.kgramLength = kgramLength;
  if (kgramsInWindow !== undefined) options.kgramsInWindow = kgramsInWindow;
  if (limitResults !== undefined) options.limitResults = limitResults;
  if (maxFingerprintCount !== undefined) options.maxFingerprintCount = maxFingerprintCount;
  if (minFragmentLength !== undefined) options.minFragmentLength = minFragmentLength;
  if (minSimilarity !== undefined) options.minSimilarity = minSimilarity;
  if (maxFingerprintPercentage !== undefined) options.maxFingerprintPercentage = maxFingerprintPercentage;
  if (input.includeComments !== undefined) {
    if (typeof input.includeComments !== "boolean") {
      throw clientError("options.includeComments must be a boolean");
    }
    options.includeComments = input.includeComments;
  }
  if (input.reportName !== undefined) {
    if (typeof input.reportName !== "string" || input.reportName.length > 128) {
      throw clientError("options.reportName must be a string of at most 128 characters");
    }
    options.reportName = input.reportName;
  }
  return options;
}

function normalizeFilename(value, index, language) {
  if (value === undefined || value === null || value === "") {
    const extension = language && LANGUAGE_EXTENSIONS.has(language)
      ? LANGUAGE_EXTENSIONS.get(language)
      : ".code";
    return `${String(index).padStart(4, "0")}${extension}`;
  }
  if (typeof value !== "string" || value.length > 255 || value.includes("\0")) {
    throw clientError(`submissions[${index}].filename is invalid`);
  }
  const basename = value.replace(/^.*[\\/]/, "").trim();
  if (!basename || basename === "." || basename === "..") {
    throw clientError(`submissions[${index}].filename is invalid`);
  }
  const safe = basename.replace(/[^a-zA-Z0-9._+\-]/g, "_").slice(0, 180);
  return `${String(index).padStart(4, "0")}-${safe || "submission"}`;
}

function normalizeSubmissions(input, language) {
  if (!Array.isArray(input) || input.length < 2) {
    throw clientError("submissions must contain at least two items");
  }
  if (input.length > MAX_SUBMISSIONS) {
    throw clientError(`submissions cannot contain more than ${MAX_SUBMISSIONS} items`);
  }

  let totalBytes = 0;
  const ids = new Set();
  return input.map((item, index) => {
    if (!item || typeof item !== "object" || Array.isArray(item)) {
      throw clientError(`submissions[${index}] must be an object`);
    }
    const id = item.id ?? item.submissionId;
    if (typeof id !== "string" && typeof id !== "number") {
      throw clientError(`submissions[${index}].id must be a string or number`);
    }
    const normalizedId = String(id).trim();
    if (!normalizedId || normalizedId.length > 128) {
      throw clientError(`submissions[${index}].id is invalid`);
    }
    if (ids.has(normalizedId)) {
      throw clientError(`submission id ${normalizedId} is duplicated`);
    }
    ids.add(normalizedId);

    const code = item.code ?? item.content ?? item.source;
    if (typeof code !== "string") {
      throw clientError(`submissions[${index}].code must be a string`);
    }
    const codeBytes = Buffer.byteLength(code, "utf8");
    if (codeBytes === 0) {
      throw clientError(`submissions[${index}].code cannot be empty`);
    }
    totalBytes += codeBytes;
    if (totalBytes > MAX_CODE_BYTES) {
      throw clientError(`combined source code exceeds ${MAX_CODE_BYTES} bytes`);
    }

    const filename = normalizeFilename(item.filename ?? item.fileName ?? item.name, index, language);
    return { id: normalizedId, filename, code };
  });
}

function createDolosFiles(submissions) {
  return submissions.map((submission, index) => new File(
    submission.filename,
    submission.code,
    undefined,
    index,
  ));
}

function regionToJson(region) {
  if (!region) {
    return null;
  }
  return {
    startRow: region.startRow,
    startCol: region.startCol,
    endRow: region.endRow,
    endCol: region.endCol,
  };
}

function pairToJson(pair, idByFileId, minFragmentLength, fragmentLimit) {
  const leftSimilarity = pair.leftTotal > 0 ? pair.leftCovered / pair.leftTotal : 0;
  const rightSimilarity = pair.rightTotal > 0 ? pair.rightCovered / pair.rightTotal : 0;
  const fragments = fragmentLimit === 0 ? [] : pair.buildFragments(minFragmentLength).slice(0, fragmentLimit).map((fragment) => ({
    left: regionToJson(fragment.leftSelection),
    right: regionToJson(fragment.rightSelection),
    kgrams: fragment.pairs.length,
  }));
  const leftId = idByFileId.get(pair.leftFile.id);
  const rightId = idByFileId.get(pair.rightFile.id);
  return {
    leftSubmissionId: leftId,
    rightSubmissionId: rightId,
    similarity: pair.similarity,
    similarityPercent: Number((pair.similarity * 100).toFixed(4)),
    // Directional coverage is useful when one submission is mostly copied
    // into a much larger submission. The left/right order is stable and maps
    // directly to the submission IDs above.
    similarityLeftToRight: leftSimilarity,
    similarityRightToLeft: rightSimilarity,
    similarity1to2: Math.round(leftSimilarity * 100),
    similarity2to1: Math.round(rightSimilarity * 100),
    totalOverlap: pair.overlap,
    longestFragment: pair.longest,
    leftCovered: pair.leftCovered,
    rightCovered: pair.rightCovered,
    leftTotal: pair.leftTotal,
    rightTotal: pair.rightTotal,
    fragments,
  };
}

async function analyzeRequest(payload) {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    throw clientError("request body must be an object");
  }
  const language = normalizeLanguage(payload.language);
  const submissions = normalizeSubmissions(payload.submissions, language);
  const inputOptions = normalizeOptions(payload.options);
  const fragmentLimit = safeInteger(payload.fragmentLimit, "fragmentLimit", { min: 0, max: MAX_FRAGMENTS }) ?? MAX_FRAGMENTS;
  const options = {
    ...inputOptions,
    ...(language ? { language } : {}),
    kgramData: false,
  };

  if (activeAnalyses >= MAX_CONCURRENT_ANALYSES) {
    const error = new Error("analysis queue is full");
    error.statusCode = 429;
    error.code = "worker_busy";
    throw error;
  }
  activeAnalyses += 1;
  const files = createDolosFiles(submissions);
  const idByFileId = new Map(files.map((file, index) => [file.id, submissions[index].id]));
  let reportPromise;
  try {
    reportPromise = new Dolos(options).analyze(files);
  } catch (error) {
    activeAnalyses -= 1;
    throw error;
  }
  // Keep the slot occupied until the underlying analysis actually finishes.
  // This matters when the request timeout fires while Tree-sitter is still busy.
  reportPromise.finally(() => { activeAnalyses -= 1; }).catch(() => {});
  let report;
  try {
    report = await withTimeout(reportPromise, ANALYSIS_TIMEOUT_MS, "analysis timed out");
  } catch (error) {
    if (!error.statusCode) {
      error.statusCode = 422;
      error.code = "analysis_failed";
    }
    throw error;
  }
  const metadata = report.metadata();
  const pairs = report.allPairs()
    .slice(0, Math.min(inputOptions.limitResults ?? MAX_PAIRS, MAX_PAIRS))
    .map((pair) => pairToJson(pair, idByFileId, inputOptions.minFragmentLength ?? 1, fragmentLimit));
  return {
    engine: {
      name: SERVICE_NAME,
      version: SERVICE_VERSION,
      package: ENGINE_PACKAGE,
      packageVersion: ENGINE_VERSION,
    },
    language: metadata.language,
    languageDetected: metadata.languageDetected,
    createdAt: metadata.createdAt,
    warnings: metadata.warnings,
    submissions: submissions.length,
    pairs,
  };
}

function withTimeout(promise, timeoutMs, message) {
  let timer;
  const timeout = new Promise((_, reject) => {
    timer = setTimeout(() => {
      const error = new Error(message);
      error.statusCode = 504;
      error.code = "analysis_timeout";
      reject(error);
    }, timeoutMs);
  });
  return Promise.race([promise, timeout]).finally(() => clearTimeout(timer));
}

async function readJson(request) {
  let bytes = 0;
  const chunks = [];
  for await (const chunk of request) {
    bytes += chunk.length;
    if (bytes > MAX_BODY_BYTES) {
      const error = new Error(`request body exceeds ${MAX_BODY_BYTES} bytes`);
      error.statusCode = 413;
      error.code = "body_too_large";
      throw error;
    }
    chunks.push(chunk);
  }
  if (chunks.length === 0) {
    throw clientError("request body is required");
  }
  try {
    return JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } catch {
    throw clientError("request body must be valid JSON");
  }
}

function sendJson(response, statusCode, body) {
  const json = JSON.stringify(body);
  response.statusCode = statusCode;
  response.setHeader("Content-Type", "application/json; charset=utf-8");
  response.setHeader("Content-Length", Buffer.byteLength(json));
  response.end(json);
}

function isAuthorized(request) {
  if (!AUTH_TOKEN) {
    return true;
  }
  const authorization = request.headers.authorization || "";
  const bearer = authorization.startsWith("Bearer ") ? authorization.slice(7).trim() : "";
  return bearer === AUTH_TOKEN || request.headers["x-worker-token"] === AUTH_TOKEN;
}

function errorResponse(error) {
  const statusCode = Number.isInteger(error?.statusCode) ? error.statusCode : 500;
  const code = error?.code || "internal_error";
  const message = statusCode >= 500 ? "Dolos analysis failed" : error.message;
  return { statusCode, body: { error: { code, message } } };
}

const server = http.createServer(async (request, response) => {
  response.setHeader("Cache-Control", "no-store");
  if (request.method === "GET" && request.url === "/healthz") {
    sendJson(response, 200, {
      ok: true,
      service: SERVICE_NAME,
      version: SERVICE_VERSION,
      engine: `${ENGINE_PACKAGE}@${ENGINE_VERSION}`,
      activeAnalyses,
      maxConcurrentAnalyses: MAX_CONCURRENT_ANALYSES,
    });
    return;
  }
  if (!isAuthorized(request)) {
    sendJson(response, 401, { error: { code: "unauthorized", message: "worker authentication required" } });
    return;
  }
  if (request.method !== "POST" || request.url !== "/v1/analyze") {
    sendJson(response, 404, { error: { code: "not_found", message: "route not found" } });
    return;
  }
  try {
    const payload = await readJson(request);
    sendJson(response, 200, await analyzeRequest(payload));
  } catch (error) {
    const result = errorResponse(error);
    sendJson(response, result.statusCode, result.body);
  }
});

server.requestTimeout = ANALYSIS_TIMEOUT_MS + 30_000;
server.headersTimeout = 15_000;
server.listen(PORT, HOST, () => {
  process.stdout.write(`${SERVICE_NAME} listening on http://${HOST}:${PORT}\n`);
});

function shutdown(signal) {
  server.close(() => process.exit(0));
  setTimeout(() => process.exit(1), 10_000).unref();
  process.stdout.write(`received ${signal}, shutting down\n`);
}

process.on("SIGTERM", () => shutdown("SIGTERM"));
process.on("SIGINT", () => shutdown("SIGINT"));
