---
description: Workflow for 95%+ Confidence, Clean, and Efficient Coding
---

# High-Confidence Coding Workflow (≥95% Solution Accuracy)

This workflow ensures that all code produced is both highly likely (>95%) to solve the intended problem, is clean and efficient, and is generated with >95% confidence, minimizing errors and adhering to linter guidelines. It is designed to be included or referenced by other workflows for critical or production-quality code.

## Phase 1: Problem Understanding & Requirements

1. **Restate the Problem Clearly**
   - Summarize the problem in your own words.
   - List all requirements, constraints, and edge cases.
   - If any ambiguity exists, halt and request clarification.

2. **Reference Source of Truth**
   - Review and cite the relevant roadmap, specs, and test files.
   - Use codebase search and web best practices as needed.

## Phase 2: Test-First & Validation Planning

3. **Write/Review Tests First**
   - Create or review tests that cover all requirements, including edge/error cases.
   - Ensure tests are behavior-focused and independent.
   - Document test coverage goals (aim for 100%, never <95%).
   - Include performance and efficiency tests where applicable.

4. **Plan Multiple Validation Layers**
   - Identify all validation steps: unit tests, integration tests, static analysis, linting, manual review, etc.
   - Specify how each will be checked before code is considered complete.
   - **Configure linters with strict rules** that enforce code cleanliness and efficiency.
   - Set up pre-commit hooks for local validation.

## Phase 3: Systematic, Modular Implementation

5. **Implement in Small, Testable Increments**
   - Make the smallest possible code change to pass the next test.
   - After each increment, run all validation steps.
   - If any validation fails, halt and fix before proceeding.
   - Ensure code is clean, efficient, and follows best practices.

6. **Self-Review and Error Analysis**
   - After each increment, review for possible errors, missed cases, or inefficiencies.
   - Run linters and fix **ALL** warnings and errors without exception.
   - Explicitly document any uncertainties or assumptions.
   - If confidence <95%, halt and request clarification or perform further validation.

## Phase 4: Confidence Checkpoint & Final Validation

7. **Confidence Checklist (Must Pass All)**
   - [ ] All tests pass (unit, integration, edge cases)
   - [ ] Code coverage ≥95%
   - [ ] Static analysis passes with zero issues
   - [ ] **ALL linter checks pass with zero warnings**
   - [ ] Solution matches all requirements and constraints
   - [ ] Code is clean, efficient, and follows best practices
   - [ ] Reasoning and implementation steps are fully documented
   - [ ] No unexplained or speculative code (no hallucination)
   - [ ] If any box is unchecked, halt and address before proceeding

8. **Peer/AI Review (Optional for Solo, Required for Team)**
   - Request a second review (human or AI) for critical code.
   - Review must focus on correctness, clarity, and confidence.
   - If confidence is <95% at any point, escalate for clarification or additional review before merging.

---
*Last Updated: 2025-05-19*

## Usage
- Reference or include this workflow in any `.windsurf/workflows` file where high-confidence, clean, and efficient code is required.
- Gather all necessary information to ensure we have >95% confidence in our solution.
- **Always run linters** and fix ALL issues before considering code complete.
- Use as a checklist for critical PRs, releases, or AI-generated code.
- Remember that no PR should be merged unless it meets ALL quality criteria with >95% confidence.