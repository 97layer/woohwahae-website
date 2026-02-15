#!/bin/bash
# 97layerOS Skills Management CLI
# Manage and execute skills from command line
#
# Usage:
#   ./scripts/skills.sh [command] [args...]
#
# Commands:
#   list              List all available skills
#   info <skill_id>   Show detailed skill information
#   execute <skill_id> [args...]  Execute a skill
#   validate <skill_id> <file>    Validate content against skill rules
#   inject <skill_ids...>         Generate prompt with skills injected
#   search <keyword>  Search for skills by keyword
#   reload           Reload skill definitions
#   help             Show this help message

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"
SKILLS_DIR="$PROJECT_ROOT/execution/core/skills"
SKILL_LOADER="$PROJECT_ROOT/system/libs/skill_loader.py"

# Change to project root
cd "$PROJECT_ROOT"

# Function to print colored output
print_color() {
    local color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Function to print header
print_header() {
    echo
    print_color "$CYAN" "╔══════════════════════════════════════════╗"
    print_color "$CYAN" "║    97layerOS Skills Management System    ║"
    print_color "$CYAN" "╚══════════════════════════════════════════╝"
    echo
}

# Check if Python is available
check_python() {
    if ! command -v python3 &> /dev/null; then
        print_color "$RED" "Error: Python 3 is not installed"
        exit 1
    fi
}

# Check if skill_loader.py exists
check_skill_loader() {
    if [ ! -f "$SKILL_LOADER" ]; then
        print_color "$RED" "Error: skill_loader.py not found at $SKILL_LOADER"
        print_color "$YELLOW" "Please ensure the Skills system is properly installed"
        exit 1
    fi
}

# List all available skills
list_skills() {
    print_header
    print_color "$BOLD" "Available Skills:"
    echo

    python3 -c "
import sys
sys.path.insert(0, '$PROJECT_ROOT')
from libs.skill_loader import SkillLoader

loader = SkillLoader()
skills = loader.list_skills()

if not skills:
    print('No skills found in $SKILLS_DIR')
else:
    for skill in skills:
        print(f\"  {skill['id']:<25} v{skill['version']:<8} {skill['description']}\")
"
}

# Show detailed information about a skill
show_skill_info() {
    local skill_id=$1

    if [ -z "$skill_id" ]; then
        print_color "$RED" "Error: Please specify a skill ID"
        echo "Usage: $0 info <skill_id>"
        exit 1
    fi

    print_header
    print_color "$BOLD" "Skill Information: $skill_id"
    echo

    python3 -c "
import sys
import json
sys.path.insert(0, '$PROJECT_ROOT')
from libs.skill_loader import SkillLoader

loader = SkillLoader()
skill = loader.load_skill('$skill_id')

if not skill:
    print('Skill not found: $skill_id')
    sys.exit(1)

print(f\"ID: {skill.get('id', 'N/A')}\")
print(f\"Name: {skill.get('name', 'N/A')}\")
print(f\"Version: {skill.get('version', 'N/A')}\")
print(f\"Description: {skill.get('description', 'N/A')}\")
print(f\"\\nRules:\")
for i, rule in enumerate(skill.get('rules', []), 1):
    print(f\"  {i}. {rule}\")
print(f\"\\nValidation Criteria:\")
for criterion in skill.get('validation_criteria', []):
    print(f\"  - {criterion}\")
"
}

# Execute a skill
execute_skill() {
    local skill_id=$1
    shift
    local args="$@"

    if [ -z "$skill_id" ]; then
        print_color "$RED" "Error: Please specify a skill ID"
        echo "Usage: $0 execute <skill_id> [args...]"
        exit 1
    fi

    print_header
    print_color "$BOLD" "Executing Skill: $skill_id"
    echo

    python3 -c "
import sys
sys.path.insert(0, '$PROJECT_ROOT')
from libs.skill_loader import SkillLoader

loader = SkillLoader()

# For now, just load and display the skill
skill = loader.load_skill('$skill_id')
if not skill:
    print('Skill not found: $skill_id')
    sys.exit(1)

print('Skill loaded successfully!')
print(f\"Arguments: $args\")
print('\\n[Note: Skill execution requires specific implementation for each skill]')
"
}

# Validate content against skill rules
validate_content() {
    local skill_id=$1
    local file_path=$2

    if [ -z "$skill_id" ] || [ -z "$file_path" ]; then
        print_color "$RED" "Error: Missing arguments"
        echo "Usage: $0 validate <skill_id> <file_path>"
        exit 1
    fi

    if [ ! -f "$file_path" ]; then
        print_color "$RED" "Error: File not found: $file_path"
        exit 1
    fi

    print_header
    print_color "$BOLD" "Validating against: $skill_id"
    echo "File: $file_path"
    echo

    python3 -c "
import sys
import json
sys.path.insert(0, '$PROJECT_ROOT')
from libs.skill_loader import SkillLoader

loader = SkillLoader()

# Read file content
with open('$file_path', 'r') as f:
    content = f.read()

# Validate
result = loader.validate_content('$skill_id', content)

if result['passed']:
    print('✅ VALIDATION PASSED')
else:
    print('❌ VALIDATION FAILED')

print(f\"\\nChecks performed: {len(result.get('checks', {}))} \")
for check, passed in result.get('checks', {}).items():
    status = '✓' if passed else '✗'
    print(f\"  {status} {check}\")

if result.get('issues'):
    print(f\"\\nIssues found ({len(result['issues'])}): \")
    for issue in result['issues']:
        print(f\"  • {issue}\")

if result.get('suggestions'):
    print(f\"\\nSuggestions:\")
    for suggestion in result['suggestions']:
        print(f\"  💡 {suggestion}\")
"
}

# Generate prompt with skills injected
inject_skills() {
    local skill_ids="$@"

    if [ -z "$skill_ids" ]; then
        print_color "$RED" "Error: Please specify at least one skill ID"
        echo "Usage: $0 inject <skill_id1> [skill_id2...]"
        exit 1
    fi

    print_header
    print_color "$BOLD" "Injecting Skills into Prompt"
    echo "Skills: $skill_ids"
    echo

    python3 -c "
import sys
sys.path.insert(0, '$PROJECT_ROOT')
from libs.skill_loader import SkillLoader

loader = SkillLoader()
skill_list = '$skill_ids'.split()

# Inject skills
enhanced_prompt = loader.inject_skills_to_prompt(skill_list)

print('Enhanced Prompt Generated:')
print('='*50)
print(enhanced_prompt)
print('='*50)
print(f\"\\nPrompt length: {len(enhanced_prompt)} characters\")
print(f\"Skills injected: {len(skill_list)}\")
"
}

# Search for skills by keyword
search_skills() {
    local keyword=$1

    if [ -z "$keyword" ]; then
        print_color "$RED" "Error: Please specify a search keyword"
        echo "Usage: $0 search <keyword>"
        exit 1
    fi

    print_header
    print_color "$BOLD" "Searching for: $keyword"
    echo

    python3 -c "
import sys
sys.path.insert(0, '$PROJECT_ROOT')
from libs.skill_loader import SkillLoader

loader = SkillLoader()
skills = loader.list_skills()

keyword = '$keyword'.lower()
matches = []

for skill in skills:
    if keyword in skill['id'].lower() or keyword in skill.get('description', '').lower():
        matches.append(skill)

if matches:
    print(f\"Found {len(matches)} matching skill(s):\")
    for skill in matches:
        print(f\"  • {skill['id']} (v{skill['version']}): {skill['description']}\")
else:
    print('No skills found matching: $keyword')
"
}

# Reload skill definitions
reload_skills() {
    print_header
    print_color "$BOLD" "Reloading Skills..."
    echo

    python3 -c "
import sys
sys.path.insert(0, '$PROJECT_ROOT')
from libs.skill_loader import SkillLoader

# Create new loader instance (forces reload)
loader = SkillLoader()
skills = loader.list_skills()

print(f\"✅ Reloaded {len(skills)} skill(s) successfully\")
for skill in skills:
    print(f\"  • {skill['id']} (v{skill['version']})\")
"
}

# Show help message
show_help() {
    print_header
    print_color "$BOLD" "Skills Management CLI"
    echo
    echo "Usage: $0 [command] [args...]"
    echo
    print_color "$BOLD" "Commands:"
    echo "  list                          List all available skills"
    echo "  info <skill_id>              Show detailed skill information"
    echo "  execute <skill_id> [args...] Execute a skill (if executable)"
    echo "  validate <skill_id> <file>   Validate content against skill rules"
    echo "  inject <skill_ids...>        Generate prompt with skills injected"
    echo "  search <keyword>             Search for skills by keyword"
    echo "  reload                       Reload skill definitions"
    echo "  help                         Show this help message"
    echo
    print_color "$BOLD" "Examples:"
    echo "  $0 list"
    echo "  $0 info brand_voice_v1"
    echo "  $0 validate brand_voice_v1 content.md"
    echo "  $0 inject brand_voice_v1 instagram_v1"
    echo "  $0 search pattern"
    echo
    print_color "$BOLD" "Available Skills:"

    # List skills inline
    if [ -d "$SKILLS_DIR" ]; then
        for skill_file in "$SKILLS_DIR"/*.skill.md; do
            if [ -f "$skill_file" ]; then
                basename "$skill_file" .skill.md | sed 's/^/  • /'
            fi
        done
    else
        echo "  [No skills directory found]"
    fi
    echo
}

# Main script logic
main() {
    # Check prerequisites
    check_python
    check_skill_loader

    # Parse command
    COMMAND="${1:-help}"
    shift || true

    case "$COMMAND" in
        list)
            list_skills
            ;;
        info)
            show_skill_info "$@"
            ;;
        execute)
            execute_skill "$@"
            ;;
        validate)
            validate_content "$@"
            ;;
        inject)
            inject_skills "$@"
            ;;
        search)
            search_skills "$@"
            ;;
        reload)
            reload_skills
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_color "$RED" "Unknown command: $COMMAND"
            echo "Use '$0 help' to see available commands"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"