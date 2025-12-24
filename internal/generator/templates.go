/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package generator

const MarkdownTemplate = `
# {{ .ResourceKind }} ({{ .Metadata.Name }})

| Property | Value |
| :--- | :--- |
| **Group** | {{ .Metadata.Group }} |
| **Scope** | {{ .Metadata.Scope }} |
| **Versions** | {{ range .Metadata.Versions }}{{ . }} {{ end }} |

## Description

{{ .Spec.Description }}

## Specification

{{ template "fields" .Spec.Fields }}

{{- define "fields" -}}
{{- range . -}}
<details>
<summary>
<code>{{ .Name }}</code>
{{ if eq .Type "string" }}<span style="color: green;">(string)</span>{{ else if eq .Type "integer" }}<span style="color: blue;">(int)</span>{{ else if eq .Type "boolean" }}<span style="color: orange;">(bool)</span>{{ else }}<b>({{ .Type }})</b>{{ end }}
{{ if .Required }}<strong>*Required*</strong>{{ end }}
</summary>

{{ if .Description }}
> {{ .Description }}
{{ end }}

{{ if or .Default .Enum }}
| Attribute | Value |
| :--- | :--- |
{{ if .Default }}| **Default** | <code>{{ .Default }}</code> |{{ end }}
{{ if .Enum }}| **Enum** | {{ range .Enum }}<code>{{ . }}</code> {{ end }} |{{ end }}
{{ end }}

{{ if .Fields }}
**Nested Fields:**
<blockquote>
{{ template "fields" .Fields }}
</blockquote>
{{ end }}

</details>
{{- end -}}
{{- end -}}
`

const HTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .ResourceKind }} ({{ .Metadata.Name }}) Documentation</title>
    <style>
        :root {
            --bg-body: #f8fafc;
            --bg-card: #ffffff;
            --text-main: #0f172a;
            --text-muted: #64748b;
            --border-color: #e2e8f0;
            --primary: #3b82f6;
            --primary-bg: #eff6ff;
            
            /* Type Colors */
            --type-string: #16a34a;
            --type-int: #0284c7;
            --type-bool: #d97706;
            --type-object: #7c3aed;
            --type-array: #db2777;
        }

        [data-theme="dark"] {
            --bg-body: #0f172a;
            --bg-card: #1e293b;
            --text-main: #f8fafc;
            --text-muted: #94a3b8;
            --border-color: #334155;
            --primary: #60a5fa;
            --primary-bg: #1e3a8a;

            --type-string: #4ade80;
            --type-int: #38bdf8;
            --type-bool: #fbbf24;
            --type-object: #a78bfa;
            --type-array: #f472b6;
        }

        * { box-sizing: border-box; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: var(--bg-body);
            color: var(--text-main);
            line-height: 1.5;
            margin: 0;
            padding: 0;
        }

        .container {
            max-width: 1000px;
            margin: 0 auto;
            padding: 2rem;
        }

        /* Header Styles */
        .doc-header {
            background: var(--bg-card);
            padding: 2rem;
            border-radius: 0.75rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            margin-bottom: 2rem;
        }

        .doc-title {
            margin: 0 0 1rem 0;
            font-size: 2rem;
            color: var(--text-main);
        }

        .meta-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 1.5rem;
            padding-bottom: 1.5rem;
            border-bottom: 1px solid var(--border-color);
        }

        .meta-item label {
            display: block;
            font-size: 0.875rem;
            color: var(--text-muted);
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .meta-item span {
            font-weight: 600;
        }

        .description {
            color: #334155;
            font-size: 1.1rem;
        }

        /* Controls */
        .controls {
            position: sticky;
            top: 0;
            z-index: 10;
            background: rgba(248, 250, 252, 0.95);
            backdrop-filter: blur(4px);
            padding: 1rem 0;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-bottom: 1px solid var(--border-color);
            margin-bottom: 2rem;
        }

        .btn-group { display: flex; gap: 0.5rem; }

        button {
            background: white;
            border: 1px solid var(--border-color);
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            cursor: pointer;
            font-weight: 500;
            color: var(--text-muted);
            transition: all 0.2s;
        }

        button:hover {
            border-color: var(--primary);
            color: var(--primary);
            background: var(--primary-bg);
        }

        #search-input {
            padding: 0.5rem 1rem;
            border: 1px solid var(--border-color);
            border-radius: 0.375rem;
            width: 250px;
            outline: none;
        }
        
        #search-input:focus {
            border-color: var(--primary);
            box-shadow: 0 0 0 3px var(--primary-bg);
        }

        /* Tree View Specification */
        .spec-container {
            background: var(--bg-card);
            border-radius: 0.75rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            overflow: hidden;
        }

        .field-row {
            border-bottom: 1px solid var(--border-color);
        }

        .field-content {
            padding: 0.75rem 1rem;
            display: flex;
            align-items: flex-start;
            gap: 0.75rem;
            transition: background 0.1s;
        }

        .field-content:hover { background-color: #f8fafc; }

        /* Indentation Lines */
        .nested-fields {
            margin-left: 1.75rem;
            border-left: 1px solid var(--border-color);
            display: none; /* Hidden by default */
        }
        
        .nested-fields.open { display: block; }

        /* Toggle Button */
        .toggle {
            background: none;
            border: none;
            padding: 0;
            width: 1.25rem;
            height: 1.25rem;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            color: var(--text-muted);
        }
        
        .toggle svg {
            transition: transform 0.2s;
            width: 14px;
            height: 14px;
        }

        .toggle.expanded svg { transform: rotate(90deg); }
        .toggle.invisible { visibility: hidden; }

        /* Field Details */
        .field-details { flex: 1; }
        
        .field-header {
            display: flex;
            align-items: baseline;
            gap: 0.75rem;
            margin-bottom: 0.25rem;
        }

        .field-name {
            font-family: "Monaco", "Menlo", monospace;
            font-weight: 600;
            font-size: 0.95rem;
            color: var(--text-main);
        }

        .field-type {
            font-size: 0.75rem;
            font-weight: 700;
            text-transform: uppercase;
            padding: 2px 6px;
            border-radius: 4px;
            background: #f1f5f9;
            color: var(--text-muted);
        }
        
        /* Type Specific Colors */
        .type-string { color: var(--type-string); background: #f0fdf4; }
        .type-integer { color: var(--type-int); background: #f0f9ff; }
        .type-boolean { color: var(--type-bool); background: #fffbeb; }
        .type-object { color: var(--type-object); background: #f5f3ff; }
        .type-array { color: var(--type-array); background: #fdf2f8; }

        .badge-req {
            font-size: 0.7rem;
            color: #dc2626;
            background: #fef2f2;
            padding: 1px 6px;
            border-radius: 99px;
            font-weight: 600;
            border: 1px solid #fee2e2;
        }

        .field-desc {
            font-size: 0.9rem;
            color: var(--text-muted);
            margin: 0.25rem 0;
        }

        .field-meta {
            font-size: 0.85rem;
            color: var(--text-muted);
            margin-top: 0.5rem;
            font-family: monospace;
        }

        .field-meta span {
            background: #f1f5f9;
            padding: 2px 6px;
            border-radius: 4px;
            margin-right: 0.5rem;
        }
        
        .hidden { display: none !important; }

        /* Dark Mode Overrides */
        [data-theme="dark"] .field-content:hover { background-color: #334155; }
        [data-theme="dark"] .field-type { background: #334155; color: #cbd5e1; }
        
        [data-theme="dark"] .type-string { background: rgba(74, 222, 128, 0.1); }
        [data-theme="dark"] .type-integer { background: rgba(56, 189, 248, 0.1); }
        [data-theme="dark"] .type-boolean { background: rgba(251, 191, 36, 0.1); }
        [data-theme="dark"] .type-object { background: rgba(167, 139, 250, 0.1); }
        [data-theme="dark"] .type-array { background: rgba(244, 114, 182, 0.1); }
        
        [data-theme="dark"] .controls { background: rgba(15, 23, 42, 0.95); }
        [data-theme="dark"] button { background: #1e293b; color: #e2e8f0; border-color: #475569; }
        [data-theme="dark"] button:hover { background: var(--primary-bg); color: var(--primary); border-color: var(--primary); }
        [data-theme="dark"] #search-input { background: #1e293b; color: white; border-color: #475569; }
    </style>
</head>
<body>

<div class="container">
    <div class="doc-header">
        <h1 class="doc-title">{{ .ResourceKind }} <span style="font-size: 0.6em; color: var(--text-muted); font-weight: normal;">{{ .Metadata.Name }}</span></h1>
        <div class="meta-grid">
            <div class="meta-item">
                <label>Group</label>
                <span>{{ .Metadata.Group }}</span>
            </div>
            <div class="meta-item">
                <label>Scope</label>
                <span>{{ .Metadata.Scope }}</span>
            </div>
            <div class="meta-item">
                <label>Versions</label>
                <span>{{ range .Metadata.Versions }}{{ . }} {{ end }}</span>
            </div>
        </div>
        <div class="description">
            {{ .Spec.Description }}
        </div>
    </div>

    <div class="controls">
        <div class="btn-group">
            <button onclick="toggleAll(true)">Expand All</button>
            <button onclick="toggleAll(false)">Collapse All</button>
            <button onclick="toggleTheme()">Theme</button>
        </div>
        <input type="text" id="search-input" placeholder="Search fields..." onkeyup="filterFields()">
    </div>

    <div class="spec-container">
        {{ template "fields" .Spec.Fields }}
    </div>
</div>

<script>
    function toggleTheme() {
        const body = document.body;
        const current = body.getAttribute('data-theme');
        const next = current === 'dark' ? 'light' : 'dark';
        body.setAttribute('data-theme', next);
        localStorage.setItem('theme', next);
    }
    
    // Init theme
    (function() {
        const saved = localStorage.getItem('theme');
        if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
            document.body.setAttribute('data-theme', 'dark');
        }
    })();

    function toggleRow(btn) {
        const row = btn.closest('.field-row');
        const nested = row.nextElementSibling; // The .nested-fields div
        
        if (nested && nested.classList.contains('nested-fields')) {
            const isHidden = getComputedStyle(nested).display === 'none';
            nested.style.display = isHidden ? 'block' : 'none';
            btn.classList.toggle('expanded');
        }
    }

    function toggleAll(expand) {
        const nestedGroups = document.querySelectorAll('.nested-fields');
        const buttons = document.querySelectorAll('.toggle:not(.invisible)');
        
        nestedGroups.forEach(el => el.style.display = expand ? 'block' : 'none');
        buttons.forEach(btn => {
            if (expand) btn.classList.add('expanded');
            else btn.classList.remove('expanded');
        });
    }

    function filterFields() {
        const query = document.getElementById('search-input').value.toLowerCase();
        const allRows = document.querySelectorAll('.field-row');
        
        // Reset if empty
        if (!query) {
            document.querySelectorAll('.field-row, .nested-fields').forEach(el => el.classList.remove('hidden'));
            // Re-apply default collapsed state logic if needed, or just leave as is
            return;
        }

        // Simple filtering: Find matches, and ensure their parents are visible
        allRows.forEach(row => {
            const name = row.querySelector('.field-name').textContent.toLowerCase();
            const desc = row.querySelector('.field-desc')?.textContent.toLowerCase() || '';
            const match = name.includes(query) || desc.includes(query);
            
            if (match) {
                row.classList.remove('hidden');
                // Walk up the DOM to reveal parents
                let parent = row.parentElement;
                while (parent && parent.classList.contains('nested-fields')) {
                    parent.style.display = 'block'; // Expand parent container
                    parent.classList.remove('hidden');
                    
                    // Find the row responsible for this nested group and show it
                    const parentTogglerRow = parent.previousElementSibling;
                    if(parentTogglerRow) {
                        parentTogglerRow.classList.remove('hidden');
                        const toggleBtn = parentTogglerRow.querySelector('.toggle');
                        if(toggleBtn) toggleBtn.classList.add('expanded');
                    }
                    parent = parent.parentElement; // Continue up
                }
            } else {
                row.classList.add('hidden');
            }
        });
    }
</script>

</body>
</html>

{{ define "fields" }}
    {{ range . }}
    <div class="field-wrapper">
        <div class="field-row">
            <div class="field-content">
                <!-- Toggle Button -->
                {{ if .Fields }}
                <button class="toggle" onclick="toggleRow(this)" aria-label="Toggle nested fields">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                        <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/>
                    </svg>
                </button>
                {{ else }}
                <button class="toggle invisible"></button>
                {{ end }}

                <div class="field-details">
                    <div class="field-header">
                        <span class="field-name">{{ .Name }}</span>
                        <span class="field-type type-{{ .Type }}">{{ .Type }}</span>
                        {{ if .Required }}<span class="badge-req">Required</span>{{ end }}
                    </div>
                    
                    {{ if .Description }}
                    <div class="field-desc">{{ .Description }}</div>
                    {{ end }}

                    {{ if or .Default .Enum }}
                    <div class="field-meta">
                        {{ if .Default }}<span>Default: {{ .Default }}</span>{{ end }}
                        {{ if .Enum }}<span>Enum: [ {{ range .Enum }}{{ . }} {{ end }}]</span>{{ end }}
                    </div>
                    {{ end }}
                </div>
            </div>
        </div>
        
        {{ if .Fields }}
        <div class="nested-fields">
            {{ template "fields" .Fields }}
        </div>
        {{ end }}
    </div>
    {{ end }}
{{ end }}
`
