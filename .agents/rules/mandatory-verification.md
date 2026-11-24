---
trigger: always_on
---

# Mandatory Verification Rule

After modifying any code, configuration, or workflow file in this repository:
1. You MUST run `npm run verify` via the terminal tool before concluding your response (follow the `verify-api` skill).
2. You MUST NOT invent ad-hoc verification methods, temporary test scripts, or one-off `curl` commands as a substitute for `npm run verify`.
3. Never mark a fix complete until `npm run verify` exits with code `0`.
