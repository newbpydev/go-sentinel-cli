---
description: Workflow for 95%+ Confidence, Low-Hallucination Coding
---

# High-Confidence Coding Workflow (≥95% Solution Accuracy)

This workflow ensures that all code produced is both highly likely (>95%) to solve the intended problem and is generated with >95% confidence, minimizing hallucinations and errors. It is designed to be included or referenced by other workflows for critical or production-quality code.

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

4. **Plan Multiple Validation Layers**
   - Identify all validation steps: unit tests, integration tests, static analysis, linting, manual review, etc.
   - Specify how each will be checked before code is considered complete.

## Phase 3: Systematic, Modular Implementation

5. **Implement in Small, Testable Increments**
   - Make the smallest possible code change to pass the next test.
   - After each increment, run all validation steps.
   - If any validation fails, halt and fix before proceeding.

6. **Self-Review and Error Analysis**
   - After each increment, review for possible errors, missed cases, or hallucinations.
   - Explicitly document any uncertainties or assumptions.
   - If confidence <95%, halt and request clarification or perform further validation.

## Phase 4: Confidence Checkpoint & Final Validation

7. **Confidence Checklist (Must Pass All)**
   - [ ] All tests pass (unit, integration, edge cases)
   - [ ] Code coverage ≥95%
   - [ ] Static analysis and linting pass
   - [ ] Solution matches all requirements and constraints
   - [ ] Reasoning and implementation steps are fully documented
   - [ ] No unexplained or speculative code (no hallucination)
   - [ ] If any box is unchecked, halt and address before proceeding

8. **Peer/AI Review (Optional for Solo, Required for Team)**
   - Request a second review (human or AI) for critical code.
   - Review must focus on correctness, clarity, and confidence.

## Phase 5: Documentation & Communication

9. **Document Everything**
   - Clearly document the solution, tests, and reasoning.
   - Note any assumptions, limitations, or open questions.

10. **Communicate Results and Next Steps**
   - Summarize what was accomplished and confidence level.
   - If confidence is <95% at any point, escalate for clarification or additional review before merging.

---
*Last Updated: 2025-05-17*

## Usage
- Reference or include this workflow in any `.windsurf/workflows` file where high-confidence, low-hallucination code is required.
- Gather all necessary information to ensure we have >95% confidence in our solution.
- Use as a checklist for critical PRs, releases, or AI-generated code.