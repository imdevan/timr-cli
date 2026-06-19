#!/usr/bin/env bash
set -euo pipefail

# Generate API documentation from Go packages using gomarkdoc
# Usage: ./docs_generate.sh [--dev]
#   --dev: Use '/' as base for local development

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGE_FILE="${ROOT_DIR}/internal/package/package.toml"
DOCS_API_DIR="${ROOT_DIR}/docs/src/content/docs/api"
DOCS_CONFIG="${ROOT_DIR}/docs/config.mjs"
DOCS_SIDEBAR="${ROOT_DIR}/docs/sidebar.mjs"

# Source shared utilities
. "${ROOT_DIR}/scripts/lib.sh"

render_flag_groups() {
  local cmd_info="$1"
  local output_file="$2"

  local num_groups=$(echo "$cmd_info" | jq '.flag_groups | length' 2>/dev/null || echo 0)
  if [ "$num_groups" -eq 0 ]; then
    return
  fi

  echo "" >>"$output_file"
  echo "## Flags" >>"$output_file"
  echo "" >>"$output_file"

  for ((g=0; g<num_groups; g++)); do
    local group_info=$(echo "$cmd_info" | jq -c ".flag_groups[$g]")
    local group_name=$(echo "$group_info" | jq -r '.name')
    local group_desc=$(echo "$group_info" | jq -r '.description')
    local group_example=$(echo "$group_info" | jq -r '.example')

    echo "### ${group_name}" >>"$output_file"
    echo "" >>"$output_file"

    if [ -n "$group_desc" ] && [ "$group_desc" != "null" ] && [ "$group_desc" != "" ]; then
      echo "$group_desc" >>"$output_file"
      echo "" >>"$output_file"
    fi

    if [ -n "$group_example" ] && [ "$group_example" != "null" ] && [ "$group_example" != "" ]; then
      echo "#### Example" >>"$output_file"
      echo "" >>"$output_file"
      echo "$group_example" >>"$output_file"
      echo "" >>"$output_file"
    fi

    local num_flags=$(echo "$group_info" | jq '.flags | length' 2>/dev/null || echo 0)
    if [ "$num_flags" -gt 0 ]; then
      echo "| Flag | Type | Description |" >>"$output_file"
      echo "|------|------|-------------|" >>"$output_file"
      for ((f=0; f<num_flags; f++)); do
        local flag_info=$(echo "$group_info" | jq -c ".flags[$f]")
        local flag_name=$(echo "$flag_info" | jq -r '.name')
        local flag_short=$(echo "$flag_info" | jq -r '.short')
        local flag_type=$(echo "$flag_info" | jq -r '.type | ascii_downcase')
        local flag_desc=$(echo "$flag_info" | jq -r '.description')

        if [ -n "$flag_short" ] && [ "$flag_short" != "null" ] && [ "$flag_short" != "" ]; then
          flag_col="-${flag_short}, --${flag_name}"
        else
          flag_col="--${flag_name}"
        fi

        echo "| \`${flag_col}\` | ${flag_type} | ${flag_desc} |" >>"$output_file"
      done
      echo "" >>"$output_file"
    fi
  done
}

echo "📦 Reading package metadata..."
PROJECT_NAME=$(parse_toml_key "$PACKAGE_FILE" "name")
CMD_DIR="${ROOT_DIR}/cmd/${PROJECT_NAME}"
DESCRIPTION=$(parse_toml_key "$PACKAGE_FILE" "description")
DOCS_SITE=$(parse_toml_key "$PACKAGE_FILE" "docs_site")
DOCS_BASE=$(parse_toml_key "$PACKAGE_FILE" "docs_base")
REPOSITORY=$(parse_toml_key "$PACKAGE_FILE" "repository")

# Use defaults if repository is empty
if [ -z "$REPOSITORY" ]; then
  REPOSITORY="https://github.com/yourusername/${PROJECT_NAME}"
fi

echo "🔧 Updating docs config..."

# Update docs/config.mjs with values from package.toml
if [ -f "$DOCS_CONFIG" ]; then
  cat >"$DOCS_CONFIG" <<EOF
const stage = process.env.NODE_ENV || "dev"
const isProduction = stage === "production"

export default {
  url: isProduction ? "$DOCS_SITE" : "http://localhost:4321",
  basePath:  isProduction ? "$DOCS_BASE" : "/",
  github: "$REPOSITORY",
  githubDocs: "$REPOSITORY",
  title: "$PROJECT_NAME",
  description: "$DESCRIPTION",
}
EOF
  echo "  ✓ Updated config.mjs with package metadata"
fi

echo "🔧 Detecting Cobra commands..."
COMMANDS_JSON=$(go run "${ROOT_DIR}/scripts/parse_commands.go" "$CMD_DIR")
num_cmds=$(echo "$COMMANDS_JSON" | jq '. | length')

echo "🔧 Generating sidebar configuration..."

# Detect commands from cmd directory
COMMANDS=""
for ((i=0; i<num_cmds; i++)); do
  cmd_info=$(echo "$COMMANDS_JSON" | jq -c ".[$i]")
  cmd_name=$(echo "$cmd_info" | jq -r '.cmd_name')
  if [ "$cmd_name" = "$PROJECT_NAME" ]; then
    continue
  fi
  COMMANDS="${COMMANDS}            { label: '${cmd_name}', link: '/commands/${cmd_name}' },
"
done

# Detect API packages
API_PACKAGES=""
if [ -d "${ROOT_DIR}/internal" ]; then
  for pkg in "${ROOT_DIR}/internal"/*/; do
    pkg_name=$(basename "$pkg")
    # Skip testutil
    if [[ "$pkg_name" == "testutil" ]]; then
      continue
    fi
    # Check if directory contains Go files (is a package, not just a folder)
    if ls "$pkg"*.go >/dev/null 2>&1; then
      API_PACKAGES="${API_PACKAGES}            { label: '${pkg_name}', link: '/api/${pkg_name}' },
"
    fi
  done
fi

# Detect API adapters
API_ADAPTERS=""
if [ -d "${ROOT_DIR}/internal/adapters" ]; then
  for adapter in "${ROOT_DIR}/internal/adapters"/*/; do
    adapter_name=$(basename "$adapter")
    API_ADAPTERS="${API_ADAPTERS}              { label: '${adapter_name}', link: '/api/adapters/${adapter_name}' },
"
  done
fi

# Check if API Reference should be included
# Include if: NODE_ENV is not "production" OR package_name is "go-cli-template"
INCLUDE_API_REFERENCE=false
if [ "${NODE_ENV:-}" != "production" ] || [ "${PROJECT_NAME}" = "go-cli-template" ]; then
  INCLUDE_API_REFERENCE=true
fi

# Build API Reference section if needed
API_REFERENCE_SECTION=""
if [ "$INCLUDE_API_REFERENCE" = true ]; then
  API_REFERENCE_SECTION="  {
    label: 'API Reference',
    items: [
${API_PACKAGES}      {
        label: 'Adapters',
        items: [
${API_ADAPTERS}        ],
      },
    ],
  },"
fi

# Conditionally include Contributing in sidebar
CONTRIBUTING_SIDEBAR=""
if [ -f "${ROOT_DIR}/CONTRIBUTING.md" ]; then
  CONTRIBUTING_SIDEBAR="sidebar.push({ label: 'Contributing', link: '/contributing' });"
fi

# Generate sidebar.mjs with dynamic environment check
cat >"$DOCS_SIDEBAR" <<EOF
import config from './config.mjs';

const apiReference = {
  label: 'API Reference',
  items: [
${API_PACKAGES}    {
      label: 'Adapters',
      items: [
${API_ADAPTERS}      ],
    },
  ],
};

const sidebar = [
  {
    label: '${PROJECT_NAME}',
    link: '/',
  },
  {
    label: 'Install',
    link: '/install',
  },
  {
    label: 'Commands',
    items: [
      { label: '${PROJECT_NAME}', link: '/commands/${PROJECT_NAME}' },
${COMMANDS}    ],
  },
  {
    label: 'Configuration',
    link: '/configuration',
  },
];

// Add API Reference if not in production or if this is go-cli-template
const isProduction = process.env.NODE_ENV === 'production';
const projectName = '${PROJECT_NAME}';
if (!isProduction || projectName === 'go-cli-template') {
  sidebar.push(apiReference);
}

${CONTRIBUTING_SIDEBAR}
export default sidebar;
EOF

echo "  ✓ Generated sidebar.mjs with detected commands and API packages"

echo "📝 Generating content pages..."

DOCS_CONTENT_DIR="docs/src/content/docs"

# Generate index page from README.md
if [ -f "README.md" ]; then
  convert_with_frontmatter "README.md" "${DOCS_CONTENT_DIR}/index.md" \
    "${PROJECT_NAME}" "${DESCRIPTION}"
  echo "  ✓ Generated index.md from README.md"
fi

# Generate install page from INSTALL.md
if [ -f "INSTALL.md" ]; then
  convert_with_frontmatter "INSTALL.md" "${DOCS_CONTENT_DIR}/install.md" \
    "Install" "Installation instructions for ${PROJECT_NAME}"
  echo "  ✓ Generated install.md from INSTALL.md"
fi

# Generate configuration page from CONFIG.md
if [ -f "CONFIG.md" ]; then
  convert_with_frontmatter "CONFIG.md" "${DOCS_CONTENT_DIR}/configuration.md" \
    "Configuration" "Configuration options for ${PROJECT_NAME}"
  echo "  ✓ Generated configuration.md from CONFIG.md"
fi

# Generate contributing page from CONTRIBUTING.md
if [ -f "CONTRIBUTING.md" ]; then
  convert_with_frontmatter "CONTRIBUTING.md" "${DOCS_CONTENT_DIR}/contributing.md" \
    "Contributing" "Contributing to ${PROJECT_NAME}"
  echo "  ✓ Generated contributing.md from CONTRIBUTING.md"
fi

# Create commands directory
mkdir -p "${DOCS_CONTENT_DIR}/commands"

# Generate root command page from root.go
if [ -f "${CMD_DIR}/root.go" ]; then
  # For root command, use the description from package.toml
  ROOT_SHORT="${DESCRIPTION}"

  # Extract root command details from parse_commands output
  root_info=$(echo "$COMMANDS_JSON" | jq -c ".[] | select(.cmd_name == \"${PROJECT_NAME}\")" 2>/dev/null || true)
  ROOT_GODOC=""
  ROOT_USE=""
  if [ -n "$root_info" ]; then
    ROOT_GODOC=$(echo "$root_info" | jq -r '.doc')
    ROOT_USE=$(echo "$root_info" | jq -r '.use')
  fi

  if [ -z "$ROOT_GODOC" ] || [ "$ROOT_GODOC" = "null" ]; then
    ROOT_GODOC="${ROOT_SHORT}"
  fi

  if [ -z "$ROOT_USE" ] || [ "$ROOT_USE" = "null" ]; then
    ROOT_USE="${PROJECT_NAME} [alias]\n${PROJECT_NAME} [command]"
  fi

  cat >"${DOCS_CONTENT_DIR}/commands/${PROJECT_NAME}.md" <<EOF
---
title: ${PROJECT_NAME}
description: ${ROOT_SHORT}
---

${ROOT_GODOC}

## Usage

\`\`\`bash
${ROOT_USE}
\`\`\`
EOF

  # Render flag groups
  render_flag_groups "$root_info" "${DOCS_CONTENT_DIR}/commands/${PROJECT_NAME}.md"

  # List all subcommands
  sub_cmds=()
  for ((i=0; i<num_cmds; i++)); do
    cmd_info=$(echo "$COMMANDS_JSON" | jq -c ".[$i]")
    cmd_name=$(echo "$cmd_info" | jq -r '.cmd_name')
    if [ "$cmd_name" = "$PROJECT_NAME" ]; then
      continue
    fi
    cmd_short=$(echo "$cmd_info" | jq -r '.short')
    if [ -z "$cmd_short" ] || [ "$cmd_short" = "null" ]; then
      cmd_short="$cmd_name"
    fi
    sub_cmds+=("- [\`${cmd_name}\`](/commands/${cmd_name}) - ${cmd_short}")
  done

  if [ "${#sub_cmds[@]}" -gt 0 ]; then
    echo "" >>"${DOCS_CONTENT_DIR}/commands/${PROJECT_NAME}.md"
    echo "## Available Commands" >>"${DOCS_CONTENT_DIR}/commands/${PROJECT_NAME}.md"
    echo "" >>"${DOCS_CONTENT_DIR}/commands/${PROJECT_NAME}.md"
    for entry in "${sub_cmds[@]}"; do
      echo "$entry" >>"${DOCS_CONTENT_DIR}/commands/${PROJECT_NAME}.md"
    done
  fi

  cat >>"${DOCS_CONTENT_DIR}/commands/${PROJECT_NAME}.md" <<EOF

## Source

See [root.go](${REPOSITORY}/blob/main/cmd/${PROJECT_NAME}/root.go) for implementation details.
EOF

  echo "  ✓ Generated commands/${PROJECT_NAME}.md"
fi

# Generate documentation for each command
for ((i=0; i<num_cmds; i++)); do
  cmd_info=$(echo "$COMMANDS_JSON" | jq -c ".[$i]")
  cmd_file=$(echo "$cmd_info" | jq -r '.go_file')
  cmd_name=$(echo "$cmd_info" | jq -r '.cmd_name')
  if [ "$cmd_name" = "$PROJECT_NAME" ]; then
    continue
  fi
  cmd_use=$(echo "$cmd_info" | jq -r '.use')
  cmd_short=$(echo "$cmd_info" | jq -r '.short')
  cmd_godoc=$(echo "$cmd_info" | jq -r '.doc')
  
  # Use cmd_name for filename and sidebar label
  cmd_url="$cmd_name"
  cmd_display="$cmd_name"

  # Use display name if Use is empty
  if [ -z "$cmd_use" ] || [ "$cmd_use" = "null" ]; then
    cmd_use="$cmd_display"
  fi

  # Use display name if Short is empty
  if [ -z "$cmd_short" ] || [ "$cmd_short" = "null" ]; then
    cmd_short="$cmd_display"
  fi

  if [ -z "$cmd_godoc" ] || [ "$cmd_godoc" = "null" ]; then
    cmd_godoc="${cmd_short}"
  fi

  # Generate command documentation
  cat >"${DOCS_CONTENT_DIR}/commands/${cmd_url}.md" <<EOF
---
title: ${cmd_display}
description: ${cmd_short}
---

${cmd_godoc}

## Usage

\`\`\`bash
${PROJECT_NAME} ${cmd_use}
\`\`\`
EOF

  # Render flag groups
  render_flag_groups "$cmd_info" "${DOCS_CONTENT_DIR}/commands/${cmd_url}.md"

  # Add source link
  cat >>"${DOCS_CONTENT_DIR}/commands/${cmd_url}.md" <<EOF

## Source

See [$(basename "$cmd_file")](${REPOSITORY}/blob/main/cmd/${PROJECT_NAME}/$(basename "$cmd_file")) for implementation details.
EOF

  echo "  ✓ Generated commands/${cmd_url}.md"
done

echo "🔧 Checking for gomarkdoc..."
if ! command -v gomarkdoc &>/dev/null; then
  echo "📦 Installing gomarkdoc..."
  go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
fi

echo "📝 Generating API documentation..."

# Generate into a temp directory to avoid Astro seeing a partially-written api/ dir
DOCS_API_TEMP="$(mktemp -d)"
trap 'rm -rf "$DOCS_API_TEMP"' EXIT

# Generate docs for each internal package
for pkg in internal/*/; do
  pkg_name=$(basename "$pkg")

  # Skip test utilities and adapters subdirectories
  if [[ "$pkg_name" == "testutil" ]]; then
    continue
  fi

  echo "  - Processing $pkg_name..."

  # Generate to temp file first
  gomarkdoc \
    --output "${DOCS_API_TEMP}/${pkg_name}.raw.md" \
    --template-file "file=${ROOT_DIR}/docs/templates/file.gotxt" \
    --footer $'## Source\n\nSee [internal/'"${pkg_name}"$'/]('"${REPOSITORY}"$'/blob/main/internal/'"${pkg_name}"$'/) for implementation details.' \
    "./$pkg" 2>/dev/null || {
    echo "    ⚠️  No exported symbols in $pkg_name"
    continue
  }

  # Add frontmatter and content (skip HTML comment and any frontmatter that gomarkdoc added)
  {
    echo "---"
    echo "title: ${pkg_name}"
    echo "description: API documentation for the ${pkg_name} package"
    echo "---"
    echo ""
    sed '1,/^# /d' "${DOCS_API_TEMP}/${pkg_name}.raw.md"
  } >"${DOCS_API_TEMP}/${pkg_name}.md"
  rm -f "${DOCS_API_TEMP}/${pkg_name}.raw.md"
done

# Generate docs for adapters
echo "  - Processing adapters..."
mkdir -p "${DOCS_API_TEMP}/adapters"

for adapter in internal/adapters/*/; do
  adapter_name=$(basename "$adapter")
  echo "    - Processing adapters/$adapter_name..."

  # Generate to temp file first
  gomarkdoc \
    --output "${DOCS_API_TEMP}/adapters/${adapter_name}.raw.md" \
    --template-file "file=${ROOT_DIR}/docs/templates/file.gotxt" \
    --footer $'## Source\n\nSee [internal/adapters/'"${adapter_name}"$'/]('"${REPOSITORY}"$'/blob/main/internal/adapters/'"${adapter_name}"$'/) for implementation details.' \
    "./$adapter" 2>/dev/null || {
    echo "      ⚠️  No exported symbols in $adapter_name"
    continue
  }

  # Add frontmatter and content (skip HTML comment and any frontmatter that gomarkdoc added)
  {
    echo "---"
    echo "title: adapters/${adapter_name}"
    echo "description: API documentation for the ${adapter_name} adapter"
    echo "---"
    echo ""
    sed '1,/^# /d' "${DOCS_API_TEMP}/adapters/${adapter_name}.raw.md"
  } >"${DOCS_API_TEMP}/adapters/${adapter_name}.md"
  rm -f "${DOCS_API_TEMP}/adapters/${adapter_name}.raw.md"
done

# Atomically swap in the newly generated api/ directory
rm -rf "$DOCS_API_DIR"
mv "$DOCS_API_TEMP" "$DOCS_API_DIR"
trap - EXIT

echo "✅ API documentation generated successfully!"
echo "📁 Output: $DOCS_API_DIR"
