FROM ghcr.io/dodona-edu/dolos-cli@sha256:39ca828cb51a0acbe7bda18bdcfe752fac29a38f5db9dc6d91f22ba6d25e6f38

ENV NODE_ENV=production \
    HOST=0.0.0.0 \
    PORT=8091 \
    MAX_CONCURRENT_ANALYSES=1

# The official image already contains Node 22, Dolos lib/core and all bundled
# Tree-sitter parsers. Copying only the worker keeps the image reproducible and
# avoids compiling native parsers a second time.
WORKDIR /repo
COPY --chown=node:node server.mjs /repo/dolos-worker/server.mjs

USER node
EXPOSE 8091
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD node -e "fetch('http://127.0.0.1:8091/healthz').then(r => { if (!r.ok) process.exit(1); }).catch(() => process.exit(1))"

ENTRYPOINT ["node", "/repo/dolos-worker/server.mjs"]
