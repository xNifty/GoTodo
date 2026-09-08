import MarkdownIt from 'markdown-it'
import type { MarkdownIt as MarkdownItInstance, StateInline, Token } from 'markdown-it'
import { isSafeImageSrc } from './taskCommentBody.ts'

export type MdNode =
  | { type: 'paragraph'; children: MdNode[] }
  | { type: 'bullet_list'; children: MdNode[] }
  | { type: 'ordered_list'; children: MdNode[] }
  | { type: 'list_item'; children: MdNode[] }
  | { type: 'strong'; children: MdNode[] }
  | { type: 'em'; children: MdNode[] }
  | { type: 'underline'; children: MdNode[] }
  | { type: 'link'; href: string; children: MdNode[] }
  | { type: 'image'; src: string; alt: string }
  | { type: 'mention'; userName: string; raw: string }
  | { type: 'task'; id: number; raw: string }
  | { type: 'text'; value: string }
  | { type: 'break' }

type MdParent = Extract<MdNode, { children: MdNode[] }>

const MENTION_RE = /^@([A-Za-z0-9_]{3,32})\b/
const BRACKET_TASK_RE = /^\[\[(\d+)\]\]/
const HASH_TASK_RE = /^#(\d+)\b/

/** http(s) links only — never javascript:, data:, or protocol-relative. */
export function isSafeHref(href: string): boolean {
  const s = href.trim()
  if (!s || /\s/.test(s) || s.includes('\\') || s.includes('<')) return false
  if (s.startsWith('/') || s.startsWith('//')) return false
  try {
    const u = new URL(s)
    return (u.protocol === 'https:' || u.protocol === 'http:') && !!u.hostname
  } catch {
    return false
  }
}

function isMentionBoundary(ch: string | undefined): boolean {
  if (ch == null || ch === '') return true
  return !/[A-Za-z0-9_@]/.test(ch)
}

function mentionRule(state: StateInline, silent: boolean): boolean {
  const pos = state.pos
  if (state.src.charCodeAt(pos) !== 0x40 /* @ */) return false
  if (!isMentionBoundary(pos > 0 ? state.src[pos - 1] : '')) return false
  const slice = state.src.slice(pos, state.posMax)
  const m = slice.match(MENTION_RE)
  if (!m) return false
  if (!silent) {
    const token = state.push('mention', 'span', 0)
    token.content = m[1]
    token.markup = m[0]
  }
  state.pos += m[0].length
  return true
}

function taskRefRule(state: StateInline, silent: boolean): boolean {
  const pos = state.pos
  const slice = state.src.slice(pos, state.posMax)
  const bracket = slice.match(BRACKET_TASK_RE)
  if (bracket) {
    if (!silent) {
      const token = state.push('task_ref', 'span', 0)
      token.content = bracket[1]
      token.markup = bracket[0]
    }
    state.pos += bracket[0].length
    return true
  }
  const hash = slice.match(HASH_TASK_RE)
  if (hash) {
    if (!silent) {
      const token = state.push('task_ref', 'span', 0)
      token.content = hash[1]
      token.markup = hash[0]
    }
    state.pos += hash[0].length
    return true
  }
  return false
}

function underlineRule(state: StateInline, silent: boolean): boolean {
  const pos = state.pos
  if (pos + 3 >= state.posMax) return false
  if (state.src.charCodeAt(pos) !== 0x2b /* + */) return false
  if (state.src.charCodeAt(pos + 1) !== 0x2b) return false

  let end = pos + 2
  while (end < state.posMax - 1) {
    if (state.src.charCodeAt(end) === 0x2b && state.src.charCodeAt(end + 1) === 0x2b) break
    end++
  }
  if (end >= state.posMax - 1) return false
  if (end === pos + 2) return false

  if (!silent) {
    state.push('underline_open', 'u', 1).markup = '++'
    const max = state.posMax
    state.pos = pos + 2
    state.posMax = end
    state.md.inline.tokenize(state)
    state.pos = end + 2
    state.posMax = max
    state.push('underline_close', 'u', -1).markup = '++'
  } else {
    state.pos = end + 2
  }
  return true
}

function commentMarkdownPlugin(md: MarkdownItInstance) {
  md.inline.ruler.before('text', 'comment_mention', mentionRule)
  md.inline.ruler.before('text', 'comment_underline', underlineRule)
  md.inline.ruler.before('link', 'comment_task_ref', taskRefRule)
}

function createParser(): MarkdownItInstance {
  const md = new MarkdownIt({
    html: false,
    breaks: true,
    linkify: false,
  })
  md.disable([
    'code',
    'fence',
    'blockquote',
    'hr',
    'html_block',
    'html_inline',
    'heading',
    'lheading',
    'reference',
    'autolink',
    'backticks',
    'strikethrough',
    'table',
    'linkify',
  ])
  md.use(commentMarkdownPlugin)
  return md
}

const parser = createParser()

function textNode(value: string): MdNode {
  return { type: 'text', value }
}

function flattenText(nodes: MdNode[]): string {
  let out = ''
  for (const n of nodes) {
    if (n.type === 'text') out += n.value
    else if (n.type === 'break') out += '\n'
    else if (n.type === 'mention') out += n.raw
    else if (n.type === 'task') out += n.raw
    else if (n.type === 'image') out += n.alt
    else if ('children' in n) out += flattenText(n.children)
  }
  return out
}

function tokensToNodes(tokens: Token[]): MdNode[] {
  const root: MdNode[] = []
  const stack: MdParent[] = []

  const current = (): MdNode[] => (stack.length ? stack[stack.length - 1].children : root)

  const pushChild = (node: MdNode) => {
    current().push(node)
  }

  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i]
    if (token.type === 'inline') {
      for (const kid of tokensToNodes(token.children || [])) pushChild(kid)
      continue
    }

    if (token.nesting === 1) {
      let node: MdParent | null = null
      switch (token.type) {
        case 'paragraph_open':
          node = { type: 'paragraph', children: [] }
          break
        case 'bullet_list_open':
          node = { type: 'bullet_list', children: [] }
          break
        case 'ordered_list_open':
          node = { type: 'ordered_list', children: [] }
          break
        case 'list_item_open':
          node = { type: 'list_item', children: [] }
          break
        case 'strong_open':
          node = { type: 'strong', children: [] }
          break
        case 'em_open':
          node = { type: 'em', children: [] }
          break
        case 'underline_open':
          node = { type: 'underline', children: [] }
          break
        case 'link_open': {
          const href = String(token.attrGet('href') || '')
          if (!isSafeHref(href)) {
            const close = tokens.findIndex((t, idx) => idx > i && t.type === 'link_close')
            const inner = close > i ? tokensToNodes(tokens.slice(i + 1, close)) : []
            for (const kid of inner) pushChild(kid)
            if (close > i) i = close
            continue
          }
          node = { type: 'link', href, children: [] }
          break
        }
        default:
          node = { type: 'paragraph', children: [] }
          break
      }
      if (!node) continue
      pushChild(node)
      stack.push(node)
      continue
    }

    if (token.nesting === -1) {
      stack.pop()
      continue
    }

    switch (token.type) {
      case 'text':
        if (token.content) pushChild(textNode(token.content))
        break
      case 'softbreak':
      case 'hardbreak':
        pushChild({ type: 'break' })
        break
      case 'mention':
        pushChild({ type: 'mention', userName: token.content, raw: token.markup || `@${token.content}` })
        break
      case 'task_ref': {
        const id = Number(token.content)
        if (id) pushChild({ type: 'task', id, raw: token.markup || `#${id}` })
        else if (token.markup) pushChild(textNode(token.markup))
        break
      }
      case 'image': {
        const src = String(token.attrGet('src') || '').trim()
        const altKids = tokensToNodes(token.children || [])
        const alt = flattenText(altKids) || String(token.content || '')
        if (isSafeImageSrc(src)) {
          pushChild({ type: 'image', src, alt })
        } else {
          pushChild(textNode(`![${alt}](${src})`))
        }
        break
      }
      default:
        if (token.content) pushChild(textNode(token.content))
        break
    }
  }

  return root
}

export function parseCommentMarkdown(body: string): MdNode[] {
  if (!body) return []
  return tokensToNodes(parser.parse(body, {}))
}

/** Walk the AST in document order, skipping wrappers. Useful in tests. */
export function flattenMdNodes(nodes: MdNode[]): MdNode[] {
  const out: MdNode[] = []
  const walk = (list: MdNode[]) => {
    for (const n of list) {
      if (
        n.type === 'paragraph' ||
        n.type === 'bullet_list' ||
        n.type === 'ordered_list' ||
        n.type === 'list_item' ||
        n.type === 'strong' ||
        n.type === 'em' ||
        n.type === 'underline' ||
        n.type === 'link'
      ) {
        walk(n.children)
      } else {
        out.push(n)
      }
    }
  }
  walk(nodes)
  return out
}
