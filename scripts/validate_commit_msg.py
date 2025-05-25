#!/usr/bin/env python3
"""
Commit Message Validator for Go Sentinel CLI

Validates commit messages according to Conventional Commits specification
and project-specific standards.

Usage: python3 scripts/validate_commit_msg.py [commit_msg_file]
"""

import sys
import re
import os
from typing import List, Tuple, Optional


class CommitMessageValidator:
    """Validates commit messages according to project standards."""

    # Conventional Commit types
    VALID_TYPES = {
        'feat': 'A new feature',
        'fix': 'A bug fix',
        'docs': 'Documentation only changes',
        'style': 'Changes that do not affect the meaning of the code',
        'refactor': 'A code change that neither fixes a bug nor adds a feature',
        'perf': 'A code change that improves performance',
        'test': 'Adding missing tests or correcting existing tests',
        'build': 'Changes that affect the build system or external dependencies',
        'ci': 'Changes to our CI configuration files and scripts',
        'chore': 'Other changes that don\'t modify src or test files',
        'revert': 'Reverts a previous commit',
        'security': 'Security improvements',
        'deps': 'Dependency updates',
        'remove': 'Remove features or code'
    }

    # Project-specific scopes
    VALID_SCOPES = {
        'cli', 'watch', 'test', 'processor', 'runner', 'cache', 'ui', 'display',
        'colors', 'icons', 'config', 'app', 'events', 'models', 'metrics',
        'complexity', 'benchmarks', 'integration', 'recovery', 'coordinator',
        'debouncer', 'watcher', 'core', 'renderer', 'docs', 'scripts', 'ci',
        'makefile', 'hooks', 'quality', 'performance', 'security', 'deps'
    }

    def __init__(self):
        self.errors: List[str] = []
        self.warnings: List[str] = []

    def validate_format(self, message: str) -> bool:
        """Validate the basic format of the commit message."""
        lines = message.strip().split('\n')
        if not lines:
            self.errors.append("Commit message cannot be empty")
            return False

        header = lines[0]

        # Check header length (recommended: ≤50 chars, max: 72 chars)
        if len(header) > 72:
            self.errors.append(f"Header too long: {len(header)} chars (max: 72)")
            return False
        elif len(header) > 50:
            self.warnings.append(f"Header long: {len(header)} chars (recommended: ≤50)")

        # Conventional Commits pattern: type(scope): description
        pattern = r'^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert|security|deps|remove)(\([a-z-]+\))?(!)?: .+'

        if not re.match(pattern, header, re.IGNORECASE):
            self.errors.append(
                f"Invalid header format: '{header}'\n"
                "Expected: type(scope): description\n"
                f"Valid types: {', '.join(self.VALID_TYPES.keys())}\n"
                f"Valid scopes: {', '.join(sorted(self.VALID_SCOPES))}"
            )
            return False

        return True

    def validate_type_and_scope(self, message: str) -> bool:
        """Validate the type and scope in the commit message."""
        header = message.strip().split('\n')[0]

        # Extract type and scope
        match = re.match(r'^([a-z]+)(\(([a-z-]+)\))?(!)?: (.+)', header, re.IGNORECASE)
        if not match:
            return False

        commit_type = match.group(1).lower()
        scope = match.group(3)
        breaking = match.group(4) == '!'
        description = match.group(5)

        # Validate type
        if commit_type not in self.VALID_TYPES:
            self.errors.append(
                f"Invalid type '{commit_type}'. Valid types: {', '.join(self.VALID_TYPES.keys())}"
            )
            return False

        # Validate scope if provided
        if scope and scope not in self.VALID_SCOPES:
            self.warnings.append(
                f"Unknown scope '{scope}'. Consider using: {', '.join(sorted(self.VALID_SCOPES))}"
            )

        # Validate description
        if not description:
            self.errors.append("Description cannot be empty")
            return False

        if description[0].isupper():
            self.warnings.append("Description should start with lowercase letter")

        if description.endswith('.'):
            self.warnings.append("Description should not end with a period")

        # Check for breaking changes
        if breaking:
            self.warnings.append("Breaking change detected - ensure CHANGELOG.md is updated")

        return True

    def validate_body_and_footer(self, message: str) -> bool:
        """Validate the body and footer of the commit message."""
        lines = message.strip().split('\n')

        if len(lines) < 2:
            return True  # No body, which is fine for simple commits

        # If there's a body, there should be a blank line after header
        if len(lines) > 1 and lines[1].strip() != '':
            self.errors.append("There must be a blank line between header and body")
            return False

        # Check body line length (recommended: ≤72 chars)
        for i, line in enumerate(lines[2:], start=3):
            if len(line) > 72:
                self.warnings.append(f"Line {i} too long: {len(line)} chars (recommended: ≤72)")

        # Check for common footers
        footer_patterns = [
            r'^Closes #\d+',
            r'^Fixes #\d+',
            r'^Resolves #\d+',
            r'^Refs #\d+',
            r'^BREAKING CHANGE:',
            r'^Co-authored-by:',
            r'^Signed-off-by:',
        ]

        for line in lines:
            for pattern in footer_patterns:
                if re.match(pattern, line):
                    # Valid footer found
                    break

        return True

    def validate_content_quality(self, message: str) -> bool:
        """Validate the quality and content of the commit message."""
        header = message.strip().split('\n')[0]

        # Extract description from after the colon
        colon_pos = header.find(': ')
        if colon_pos == -1:
            return False

        description = header[colon_pos + 2:].strip()

        # Check for meaningful description
        vague_words = ['fix', 'update', 'change', 'modify', 'improve', 'refactor', 'cleanup']
        if any(description.lower().strip() == word for word in vague_words):
            self.warnings.append(
                f"Description '{description}' is too vague. Be more specific about what was changed."
            )

        # Check minimum description length
        if len(description) < 10:
            self.warnings.append("Description should be at least 10 characters for clarity")

        # Check for imperative mood (should start with verb)
        imperative_indicators = [
            'add', 'fix', 'remove', 'update', 'improve', 'refactor', 'implement',
            'create', 'delete', 'enhance', 'optimize', 'clean', 'extract',
            'rename', 'move', 'split', 'merge', 'integrate', 'configure',
            'resolve', 'install', 'setup', 'build', 'deploy', 'release'
        ]

        first_word = description.split()[0].lower() if description.split() else ''
        if first_word not in imperative_indicators:
            self.warnings.append(
                f"Consider using imperative mood: '{description}' → 'add ...', 'fix ...', etc."
            )

        return True

    def validate_commit_message(self, message: str) -> Tuple[bool, List[str], List[str]]:
        """Validate the entire commit message."""
        self.errors = []
        self.warnings = []

        if not message or not message.strip():
            self.errors.append("Commit message cannot be empty")
            return False, self.errors, self.warnings

        # Run all validations
        format_ok = self.validate_format(message)
        type_scope_ok = self.validate_type_and_scope(message) if format_ok else False
        body_footer_ok = self.validate_body_and_footer(message) if format_ok else False
        quality_ok = self.validate_content_quality(message) if format_ok else False

        is_valid = format_ok and type_scope_ok and body_footer_ok and quality_ok

        return is_valid, self.errors, self.warnings


def read_commit_message(file_path: Optional[str] = None) -> str:
    """Read commit message from file or stdin."""
    if file_path and os.path.exists(file_path):
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    elif not sys.stdin.isatty():
        return sys.stdin.read()
    else:
        # For testing, read from command line argument
        return ' '.join(sys.argv[1:]) if len(sys.argv) > 1 else ''


def print_help():
    """Print help information about commit message format."""
    help_text = """
🔍 Commit Message Format Guide

Format: type(scope): description

Valid Types:
  feat     - New feature
  fix      - Bug fix
  docs     - Documentation changes
  style    - Code style changes (formatting, etc.)
  refactor - Code refactoring
  perf     - Performance improvements
  test     - Adding or updating tests
  build    - Build system or dependency changes
  ci       - CI/CD configuration changes
  chore    - Other changes (tooling, etc.)
  security - Security improvements
  deps     - Dependency updates

Common Scopes:
  cli, watch, test, processor, runner, cache, ui, display,
  colors, config, app, events, models, metrics, complexity

Examples:
  ✅ feat(cli): add complexity analysis command
  ✅ fix(watch): resolve file debouncing race condition
  ✅ docs(readme): update installation instructions
  ✅ test(processor): add integration tests for JSON parsing
  ✅ refactor(ui): extract color formatting to separate package

Breaking Changes:
  feat(api)!: remove deprecated endpoints

Body (optional):
  - Use imperative mood
  - Explain why, not what
  - Keep lines ≤72 characters

Footer (optional):
  Closes #123
  BREAKING CHANGE: API endpoints changed
"""
    print(help_text)


def main():
    """Main entry point for commit message validation."""

    # Check for help flag
    if '--help' in sys.argv or '-h' in sys.argv:
        print_help()
        sys.exit(0)

    # Read commit message
    commit_msg_file = sys.argv[1] if len(sys.argv) > 1 else None
    message = read_commit_message(commit_msg_file)

    if not message.strip():
        print("❌ Error: No commit message provided")
        print("\nUse --help for format guide")
        sys.exit(1)

    # Validate commit message
    validator = CommitMessageValidator()
    is_valid, errors, warnings = validator.validate_commit_message(message)

    # Print results
    if is_valid and not warnings:
        print("✅ Commit message format is valid")
        sys.exit(0)

    print("📝 Commit Message Validation Results")
    print("=" * 50)

    if errors:
        print("\n❌ ERRORS (must fix):")
        for error in errors:
            print(f"  • {error}")

    if warnings:
        print("\n⚠️  WARNINGS (recommended fixes):")
        for warning in warnings:
            print(f"  • {warning}")

    if not is_valid:
        print(f"\n❌ Commit message validation failed")
        print("\nUse --help for format guide")
        sys.exit(1)
    else:
        print(f"\n✅ Commit message is valid (with warnings)")
        sys.exit(0)


if __name__ == '__main__':
    main()
