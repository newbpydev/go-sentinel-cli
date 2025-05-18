---
description: Systematic workflow for fixing failing tests using TDD principles
---

> **This workflow incorporates the [High-Confidence Coding Workflow](high-confidence-coding.md) to ensure all fixes are ≥95% likely to solve the problem, with systematic validation and minimal hallucination.**

# Workflow: Fixing Failing Tests (TDD-Driven)

This workflow ensures a systematic, repeatable, and thorough process for fixing failing tests in a TDD environment. It first addresses any file-level errors before moving on to test failures. The workflow leverages project files, codebase search, and web best practices. **Never skip a step.**

## Phase 1: File-Level Error Resolution

### Step 1: Identify File Errors
- Run the TypeScript compiler or linter to identify any syntax or type errors in the file.
- Check for any runtime errors that might occur when the file is loaded.
- Note all errors with their exact locations and messages.

### Step 2: Fix File Errors
- Address syntax errors first (missing brackets, semicolons, etc.).
- Fix type errors by ensuring all variables and functions have proper type annotations.
- Resolve any import/export issues.
- Ensure the file can be successfully compiled/transpiled without errors.

### Step 3: Verify File Integrity
- Run any file-specific validation or compilation steps.
- Ensure the file follows project coding standards.
- Verify that the file can be imported/required without errors.

## Phase 2: Test-First Development

### Step 4: Identify and Document the Failing Test
- Run the full test suite to identify all failing tests.
- Note the exact test(s) failing, error messages, and stack traces.
- Update `CHANGES.md` and/or a dedicated bug tracker with a description of the failure.

### Step 5: Analyze Root Cause
- Read the test code and related implementation code.
- Use codebase search (`grep`, IDE search, etc.) to find all relevant usages and dependencies.
- If the root cause is unclear, search the web for similar issues or best practices.
- If the failure is flaky, rerun the test to confirm reproducibility.

### Step 6: Minimize Scope of Fix
- Isolate the minimal code change required to fix the test while avoiding regressions.
- Write (or update) a test that precisely captures the correct behavior, if not already present.
- Ensure the test is independent, repeatable, and focused on behavior (not implementation).

### Step 7: Implement the Fix
- Make the minimal code change needed to pass the failing test.
- Follow project coding standards and keep changes as small as possible.
- Add comments referencing the test and root cause if appropriate.

### Step 8: Run All Tests
- Run the entire test suite (not just the fixed test) to ensure no regressions are introduced.
- If new failures appear, analyze and address them before proceeding.

## Phase 3: Code Quality and Documentation

> **Perform a high-confidence checkpoint before considering the fix complete. Review all steps in [high-confidence-coding.md]:**
> - Complete the confidence checklist
> - Ensure ≥95% test coverage and all validations pass
> - Document reasoning, edge cases, and uncertainties
> - If confidence is <95%, halt and request clarification or peer review before merging

### Step 9: Refactor with Confidence
- If the fix introduces code smells or duplication, refactor the affected code.
- Ensure all tests still pass after refactoring.
- Run static analysis tools to catch any new issues.

### Step 10: Update Documentation
- Update `CHANGES.md` with details of the fix and affected tests.
- Mark the relevant item as complete or in-progress in `ROADMAP.md`.
- Add or update code comments to clarify tricky logic.
- Ensure JSDoc/TSDoc comments are accurate and complete.

### Step 11: Peer Review (if applicable)
- Request a code review or pair programming session to validate the fix and tests.
- Incorporate feedback and re-run tests as needed.

### Step 12: Final Verification
- Confirm that all tests pass in CI and locally.
- Verify that all file-level errors remain resolved.
- Ensure the fix is documented and the roadmap is up to date.

### Step 13: Close the Loop
- Mark the issue as resolved in the tracker.
- Communicate the fix to relevant stakeholders if needed.
- Consider adding regression tests if the issue was particularly subtle or impactful.

---

### Tips & Best Practices
- **Fix File Errors First:** Always resolve file-level errors before addressing test failures.
- **Start Small:** Break fixes into the smallest possible increments.
- **Test Behavior, Not Implementation:** Focus on what the code should do, not how it does it.
- **Integrate Regularly:** Use CI to catch integration issues early.
- **Collaborate:** Use code reviews and pair programming for knowledge sharing.
- **Leverage Tools:** Use codebase search, web research, and project documentation to inform your fix.
- **Documentation is Key:** Keep documentation in sync with code changes.

---

_This workflow is aligned with TDD and project best practices. For response formatting and reporting, follow the structure in `prefered-response.md`._

Gather all necessary information to ensure we have >95% confidence in our solution.