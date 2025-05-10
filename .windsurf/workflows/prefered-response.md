---
description: preferred response for most cases
---

## Core Response Principles

1. **TDD First, Always**: Begin every implementation task by writing failing tests that clearly define expected behavior. Tests must be written before ANY code implementation.

2. **Systematic, Step-by-Step Process**: @web @ROADMAP.md Work methodically through tasks in sequence, respecting dependencies. Never skip ahead to later tasks before completing prerequisites.

3. **Roadmap-Driven Development**: The ROADMAP.md file is our source of truth. Never deviate from it @ROADMAP.md - if changes are needed, update the roadmap first, then implement.

4. **Comprehensive Testing**: Cover all edge cases, error conditions, and happy paths. Every feature must have exhaustive test coverage. Identify potential weak points in tests.

5. **Clear Progress Tracking**: Always update completed tasks in the roadmap with checkmarks. Provide visual progress indicators in summaries.

## Response Structure

When completing tasks or providing updates, use this structured format:

### 1. Summary Header
Brief statement of what was accomplished, using action verbs and specific outcomes.

### 2. Completed Work (✅)
Detailed list of completed items with checkmarks, grouped by categories:
- **Tests Implemented**: What test cases were written
- **Code Implemented**: What functionality was delivered
- **Coverage Metrics**: Test coverage percentages (if available)
- **Roadmap Updates**: Which roadmap items were marked complete

### 3. Technical Details
- Design patterns used
- Architecture decisions made
- Performance considerations
- Security implications addressed

### 4. Next Steps
- Clear prioritized list of upcoming tasks from the roadmap
- Any potential challenges to consider
- Suggested approach for the next implementation

### 5. Questions/Decisions
Only if needed, specific questions requiring user input or decisions

Example Output from a project called BubblyUI after the tests have passed, and the Roadmap and CHANGES files habe been updated and checked for the next tasks:

Great work! We've successfully implemented the event type system for the BubblyUI framework following the test-driven development approach. All tests are passing, and I've updated the ROADMAP.md to reflect our progress.

What We've Accomplished
We've completed the first subtask of Phase 4 Task 2 (Event Handling):

✅ Create event type system

✅ Designed generic Event interface
✅ Implemented standard event types (keyboard, mouse, window size)
✅ Created custom event type factory
✅ Added event metadata structure
✅ Designed event type hierarchy/categorization

Our implementation includes:

A comprehensive Event interface with methods for event propagation control
Standard event wrappers for Bubble Tea messages
Support for custom user-defined events
Event categorization and metadata

Next Steps
Should we create more edge case tests?