export type CommentBodyPart =
  | { type: 'text'; value: string }
  | { type: 'task'; id: number; raw: string }

const TASK_REF_RE = /\[\[(\d+)\]\]|#(\d+)\b/g

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
  const re = new RegExp(TASK_REF_RE.source, 'g')
  let last = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(body))) {
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
