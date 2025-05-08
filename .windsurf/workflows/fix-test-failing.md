---
description: Systematic workflow for fixing failing tests using TDD principles
---

# Workflow: Fixing Failing Tests (TDD-Driven)

This workflow ensures a systematic, repeatable, and thorough process for fixing failing tests in a TDD environment. It leverages project files, codebase search, and web best practices. **Never skip a step.**

## Step 1: Identify and Document the Failing Test
- Run the full test suite to identify all failing tests.
- Note the exact test(s) failing, error messages, and stack traces.
- Update `CHANGES.md` and/or a dedicated bug tracker with a description of the failure.

## Step 2: Analyze Root Cause
- Read the test code and related implementation code.
- Use codebase search (`grep`, IDE search, etc.) to find all relevant usages and dependencies.
- If the root cause is unclear, search the web for similar issues or best practices.
- If the failure is flaky, rerun the test to confirm reproducibility (see "auto-restarts" best practice).

## Step 3: Minimize Scope of Fix
- Isolate the minimal code change required to fix the test while avoiding regressions.
- Write (or update) a test that precisely captures the correct behavior, if not already present.
- Ensure the test is independent, repeatable, and focused on behavior (not implementation).

## Step 4: Implement the Fix
- Make the minimal code change needed to pass the failing test.
- Follow project coding standards and keep changes as small as possible.
- Add comments referencing the test and root cause if appropriate.

## Step 5: Run All Tests
- Run the entire test suite (not just the fixed test) to ensure no regressions are introduced.
- If new failures appear, analyze and address them before proceeding.

## Step 6: Refactor with Confidence
- If the fix introduces code smells or duplication, refactor the affected code.
- Ensure all tests still pass after refactoring.

## Step 7: Update Documentation and Roadmap
- Update `CHANGES.md` with details of the fix and affected tests.
- Mark the relevant item as complete or in-progress in `ROADMAP.md`.
- Add or update code comments to clarify tricky logic.

## Step 8: Peer Review (if applicable)
- Request a code review or pair programming session to validate the fix and tests.
- Incorporate feedback and re-run tests as needed.

## Step 9: Final Verification
- Confirm that all tests pass in CI and locally.
- Ensure the fix is documented and the roadmap is up to date.

## Step 10: Close the Loop
- Mark the issue as resolved in the tracker.
- Communicate the fix to relevant stakeholders if needed.

---

### Tips & Best Practices
- **Start Small:** Break fixes into the smallest possible increments.
- **Test Behavior, Not Implementation:** Focus on what the code should do.
- **Integrate Regularly:** Use CI to catch integration issues early.
- **Collaborate:** Use code reviews and pair programming for knowledge sharing.
- **Leverage Tools:** Use codebase search, web research, and project documentation to inform your fix.

---

_This workflow is aligned with TDD and project best practices. For response formatting and reporting, follow the structure in `prefered-response.md`._
