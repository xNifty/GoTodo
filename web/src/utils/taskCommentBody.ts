export type CommentBodyPart =
  | { type: 'text'; value: string }
  | { type: 'task'; id: number; raw: string }
  | { type: 'mention'; userName: string; raw: string }

export type MentionToken = {
  start: number
  end: number
  query: string
}

const TASK_REF_RE = /\[\[(\d+)\]\]|#(\d+)\b/g
const COMMENT_TOKEN_RE =
  /\[\[(\d+)\]\]|#(\d+)\b|(^|[^A-Za-z0-9_@])@([A-Za-z0-9_]{3,32})\b/g
const MENTION_RE = /(^|[^A-Za-z0-9_@])@([A-Za-z0-9_]{3,32})\b/g

export function extractTaskRefIDs(body: string): number[] {
  const ids: number[] = []
  const seen = new Set<number>()
  const re = new RegExp(TASK_REF_RE.source, 'g')
  let m: RegExpExecArray | null
  while ((m = re.exec(body))) {
    const id = Number(m[1] || m[2])
    if (!id || seen.has(id)) continue
    seen.add(id)
    ids.push(id)
  }
  return ids
}

export function extractMentionNames(body: string): string[] {
  const names: string[] = []
  const seen = new Set<string>()
  const re = new RegExp(MENTION_RE.source, 'g')
  let m: RegExpExecArray | null
  while ((m = re.exec(body))) {
    const name = m[2]
    const key = name.toLowerCase()
    if (!name || seen.has(key)) continue
    seen.add(key)
    names.push(name)
  }
  return names
}

export function mentionTokenAtCursor(body: string, cursor: number): MentionToken | null {
  if (cursor < 0) cursor = 0
  if (cursor > body.length) cursor = body.length
  const before = body.slice(0, cursor)
  const m = before.match(/(?:^|[^A-Za-z0-9_@])@([A-Za-z0-9_]{0,32})$/)
  if (!m) return null
  const queryBefore = m[1]
  const afterMatch = body.slice(cursor).match(/^([A-Za-z0-9_]{0,32})/)
  const queryAfter = afterMatch?.[1] ?? ''
  if (queryBefore.length + queryAfter.length > 32) return null
  const atIndex = before.length - 1 - queryBefore.length
  if (atIndex < 0 || body[atIndex] !== '@') return null
  return {
    start: atIndex,
    end: cursor + queryAfter.length,
    query: queryBefore + queryAfter,
  }
}

export function insertMention(
  body: string,
  token: MentionToken,
  userName: string,
): { body: string; cursor: number } {
  const mention = `@${userName}`
  const after = body.slice(token.end)
  const spacer = /^\s/.test(after) ? '' : ' '
  const next = body.slice(0, token.start) + mention + spacer + after
  return { body: next, cursor: token.start + mention.length + spacer.length }
}

export function isInsertedTaskRef(body: string, id: number): boolean {
  return new RegExp(`\\[\\[${id}\\]\\]`).test(body)
}

export function insertTaskRef(body: string, id: number): string {
  if (isInsertedTaskRef(body, id)) return body
  const hash = new RegExp(`#${id}\\b`)
  if (hash.test(body)) return body.replace(hash, `[[${id}]]`)
  const trimmed = body.trimEnd()
  if (!trimmed) return `[[${id}]]`
  return `${trimmed} [[${id}]]`
}

export function splitCommentBody(body: string): CommentBodyPart[] {
  const parts: CommentBodyPart[] = []
  const re = new RegExp(COMMENT_TOKEN_RE.source, 'g')
  let last = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(body))) {
    const mentionName = m[4]
    if (mentionName) {
      const prefix = m[3] ?? ''
      const tokenStart = m.index + prefix.length
      if (tokenStart > last) {
        parts.push({ type: 'text', value: body.slice(last, tokenStart) })
      }
      parts.push({ type: 'mention', userName: mentionName, raw: `@${mentionName}` })
      last = m.index + m[0].length
      continue
    }
    if (m.index > last) {
      parts.push({ type: 'text', value: body.slice(last, m.index) })
    }
    const id = Number(m[1] || m[2])
    parts.push({ type: 'task', id, raw: m[0] })
    last = m.index + m[0].length
  }
  if (last < body.length) {
    parts.push({ type: 'text', value: body.slice(last) })
  }
  return parts
}
