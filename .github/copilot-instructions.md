# Copilot Instructions

## Scope
These instructions apply to the entire repository.

## Code Review Rules
- Prioritize correctness, security, and regression risk over style-only comments.
- Report findings in severity order: critical, high, medium, low.
- Include concrete file references and actionable fixes for each finding.
- If no major issues are found, explicitly state that and list any residual risks.
- Keep summaries short; findings come first.
- Always return review feedback in Japanese.
- For Gin handlers and middleware, check request binding, validation, and error responses carefully.
- Verify handlers stop execution after `Abort`, `AbortWithStatus`, or `AbortWithStatusJSON` when required.
- Flag cases where HTTP status codes, response bodies, or content types do not match the API behavior.
- Check middleware order for security and behavior regressions, especially recovery, logging, auth, and CORS.
- Review uses of `gin.Context` for unsafe assumptions about query params, path params, headers, and body data.
- Call out security risks such as missing input validation, unbounded request bodies, leaking internal errors, or unsafe proxy/trusted header handling.

## Lint Rules
- Run or assume standard Go linting practices before suggesting code is complete.
- Prefer fixes that satisfy common linters without suppressing warnings.
- Avoid introducing unnecessary complexity just to silence lint.
- If lint cannot be run, clearly state that limitation and why.
- Prefer patterns that common Go linters accept in Gin code, including explicit error handling for `ShouldBind*` and helper return values.
- Avoid ignored errors in request parsing, JSON rendering, server startup, and shutdown paths.
- Do not introduce unreachable branches or duplicated response writes in handlers and middleware.
- Keep handler and middleware helpers small enough that control flow and abort behavior remain obvious under lint review.

## Format Rules
- Ensure Go code is formatted with `gofmt` conventions.
- Keep imports organized in standard Go format.
- Do not leave formatting-only noise in unrelated files.
- Keep Gin route, handler, and middleware definitions formatted for readability, especially when chaining groups or nested routes.
- Prefer consistent JSON response shapes and field naming within the same API surface.
- Keep binding and validation tags readable and aligned with existing struct tag style in the file.

## Documentation Rules
- When making any code changes (new features, refactoring, route additions, structural changes), always update the following two files:
  - `README.md`: Keep the tech stack versions and run instructions up to date.
  - `docs/architecture.md`: Keep the directory structure, routing table, and design policy sections accurate and current.
- Do not skip this step even for small changes. Documentation must stay in sync with the code at all times.

- Before creating or finalizing any commit, review the staged or proposed diff against the Code Review Rules and report findings first.
- Do not proceed directly to commit creation when a review has not been performed in the current task.
- If the user asks for a commit, perform the review first and only then propose or create the commit when no blocking issues remain.
- Before finalizing a commit message, verify it follows Conventional Commits.
- Use the pattern: `<type>(<scope>): <subject>` where useful.
- Prefer concise, imperative subjects.
- Common types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`.
- If the message does not conform, propose a corrected Conventional Commits message.
