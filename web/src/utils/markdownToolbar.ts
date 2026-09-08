export type ToolbarRange = {
  body: string
  start: number
  end: number
}

function clampRange(body: string, start: number, end: number): { start: number; end: number } {
  let s = start
  let e = end
  if (s < 0) s = 0
  if (e < s) e = s
  if (s > body.length) s = body.length
  if (e > body.length) e = body.length
  return { start: s, end: e }
}

function lineBounds(body: string, start: number, end: number): { start: number; end: number } {
  const lineStart = body.lastIndexOf('\n', start - 1) + 1
  const nl = body.indexOf('\n', end)
  const lineEnd = nl === -1 ? body.length : nl
  return { start: lineStart, end: lineEnd }
}

/** Wrap the selection with a marker (`**`, `*`, `++`). Toggles if already wrapped. */
export function wrapInline(
  body: string,
  start: number,
  end: number,
  marker: string,
  placeholder = 'text',
): ToolbarRange {
  const range = clampRange(body, start, end)
  const before = body.slice(0, range.start)
  const selected = body.slice(range.start, range.end)
  const after = body.slice(range.end)

  if (
    selected.startsWith(marker) &&
    selected.endsWith(marker) &&
    selected.length >= marker.length * 2
  ) {
    const inner = selected.slice(marker.length, selected.length - marker.length)
    const next = before + inner + after
    return { body: next, start: range.start, end: range.start + inner.length }
  }

  if (before.endsWith(marker) && after.startsWith(marker)) {
    const next = before.slice(0, before.length - marker.length) + selected + after.slice(marker.length)
    return {
      body: next,
      start: range.start - marker.length,
      end: range.end - marker.length,
    }
  }

  const inner = selected || placeholder
  const wrapped = `${marker}${inner}${marker}`
  const next = before + wrapped + after
  if (selected) {
    const cursor = range.start + wrapped.length
    return { body: next, start: cursor, end: cursor }
  }
  return {
    body: next,
    start: range.start + marker.length,
    end: range.start + marker.length + inner.length,
  }
}

const UL_RE = /^(\s*)- /
const OL_RE = /^(\s*)\d+\. /

/** Prefix selected lines with `- ` or `1. `. Toggles an existing list of the same kind. */
export function prefixLines(
  body: string,
  start: number,
  end: number,
  kind: 'ul' | 'ol',
): ToolbarRange {
  const range = clampRange(body, start, end)
  const bounds = lineBounds(body, range.start, range.end)
  const block = body.slice(bounds.start, bounds.end)
  const lines = block.length || range.start !== range.end ? block.split('\n') : ['']

  const allListed =
    lines.length > 0 &&
    lines.every((line) => (kind === 'ul' ? UL_RE.test(line) : OL_RE.test(line)))

  const nextLines = lines.map((line, i) => {
    if (kind === 'ul') {
      if (allListed) return line.replace(UL_RE, '$1')
      if (OL_RE.test(line)) return line.replace(OL_RE, '$1- ')
      return `- ${line}`
    }
    if (allListed) return line.replace(OL_RE, '$1')
    if (UL_RE.test(line)) return line.replace(UL_RE, `$1${i + 1}. `)
    return `${i + 1}. ${line}`
  })

  const replacement = nextLines.join('\n')
  const next = body.slice(0, bounds.start) + replacement + body.slice(bounds.end)
  return {
    body: next,
    start: bounds.start,
    end: bounds.start + replacement.length,
  }
}

function escapeLinkText(text: string): string {
  return text.replace(/[[\]]/g, '')
}

/** Insert `[title](href)` at the selection, replacing any selected text. */
export function insertLinkMarkup(
  body: string,
  start: number,
  end: number,
  title: string,
  href: string,
): ToolbarRange {
  const range = clampRange(body, start, end)
  const label = escapeLinkText(title.trim() || href)
  const inserted = `[${label}](${href})`
  const next = body.slice(0, range.start) + inserted + body.slice(range.end)
  const cursor = range.start + inserted.length
  return { body: next, start: cursor, end: cursor }
}

const UL_ITEM_RE = /^(\s*)([-*+]) (.*)$/
const OL_ITEM_RE = /^(\s*)(\d+)\. (.*)$/

/**
 * GitHub-style list continuation: Enter on a list item starts the next item.
 * Enter on an empty list item leaves the list.
 */
export function continueListOnEnter(body: string, cursor: number): ToolbarRange | null {
  const pos = clampRange(body, cursor, cursor).start
  const lineStart = body.lastIndexOf('\n', pos - 1) + 1
  const nl = body.indexOf('\n', pos)
  const lineEnd = nl === -1 ? body.length : nl
  const line = body.slice(lineStart, lineEnd)
  const ol = line.match(OL_ITEM_RE)
  const ul = line.match(UL_ITEM_RE)
  if (!ol && !ul) return null

  const indent = ol ? ol[1] : ul![1]
  const content = ol ? ol[3] : ul![3]
  const marker = ol ? `${ol[2]}. ` : `${ul![2]} `
  const contentStart = lineStart + indent.length + marker.length
  if (pos < contentStart) return null

  if (!content.trim()) {
    const next = body.slice(0, lineStart) + body.slice(lineEnd)
    return { body: next, start: lineStart, end: lineStart }
  }

  const nextMarker = ol ? `${Number(ol[2]) + 1}. ` : `${ul![2]} `
  const insert = `\n${indent}${nextMarker}`
  const next = body.slice(0, pos) + insert + body.slice(pos)
  const nextCursor = pos + insert.length
  return { body: next, start: nextCursor, end: nextCursor }
}

/** Prepend https:// when the scheme is missing. Returns null when the URL is not http(s). */
export function normalizeLinkHref(raw: string): string | null {
  let s = raw.trim()
  if (!s) return null
  if (!/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(s)) s = `https://${s}`
  try {
    const u = new URL(s)
    if ((u.protocol !== 'https:' && u.protocol !== 'http:') || !u.hostname) return null
    return u.href
  } catch {
    return null
  }
}

export function wouldExceedLimit(nextLength: number, maxLength: number): boolean {
  return maxLength > 0 && nextLength > maxLength
}
