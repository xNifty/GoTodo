import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { flattenMdNodes, isSafeHref, parseCommentMarkdown } from './commentMarkdown.ts'

function typesOf(body: string): string[] {
  return flattenMdNodes(parseCommentMarkdown(body)).map((n) => n.type)
}

describe('isSafeHref', () => {
  it('allows http and https', () => {
    assert.equal(isSafeHref('https://example.com/a'), true)
    assert.equal(isSafeHref('http://example.com/a'), true)
  })
  it('rejects javascript, data, and relative tricks', () => {
    assert.equal(isSafeHref('javascript:alert(1)'), false)
    assert.equal(isSafeHref('data:text/html,hi'), false)
    assert.equal(isSafeHref('//evil.example/x'), false)
    assert.equal(isSafeHref('/internal'), false)
  })
})

describe('parseCommentMarkdown existing tokens', () => {
  it('keeps mentions and task refs working', () => {
    assert.deepEqual(typesOf('Hi @Ada see #12 and [[9]]'), ['text', 'mention', 'text', 'task', 'text', 'task'])
  })

  it('does not italicize underscores in mentions', () => {
    const nodes = flattenMdNodes(parseCommentMarkdown('ping @Bob_Editor please'))
    const mention = nodes.find((n) => n.type === 'mention')
    assert.equal(mention?.type, 'mention')
    if (mention?.type === 'mention') {
      assert.equal(mention.userName, 'Bob_Editor')
      assert.equal(mention.raw, '@Bob_Editor')
    }
    assert.equal(nodes.some((n) => n.type === 'em'), false)
  })

  it('renders inserted markdown as an image part', () => {
    const nodes = flattenMdNodes(parseCommentMarkdown('see ![cat](https://cdn.example.com/cat.png) please'))
    const image = nodes.find((n) => n.type === 'image')
    assert.equal(image?.type, 'image')
    if (image?.type === 'image') {
      assert.equal(image.alt, 'cat')
      assert.equal(image.src, 'https://cdn.example.com/cat.png')
    }
  })

  it('does not treat unsafe URLs as images', () => {
    const nodes = flattenMdNodes(parseCommentMarkdown('![x](javascript:alert(1))'))
    assert.equal(nodes.some((n) => n.type === 'image'), false)
  })

  it('does not eat a following task number out of a URL', () => {
    const nodes = flattenMdNodes(parseCommentMarkdown('![shot](https://cdn.example.com/a.png)\n#42'))
    assert.equal(nodes[0].type, 'image')
    assert.equal(nodes[nodes.length - 1].type, 'task')
  })

  it('does not treat #123 as a heading', () => {
    const nodes = flattenMdNodes(parseCommentMarkdown('#123'))
    assert.equal(nodes.length, 1)
    assert.equal(nodes[0].type, 'task')
    if (nodes[0].type === 'task') assert.equal(nodes[0].id, 123)
  })
})

describe('parseCommentMarkdown formatting', () => {
  it('parses bold, italic, and underline', () => {
    const ast = parseCommentMarkdown('**bold** *italic* ++under++')
    assert.equal(ast[0].type, 'paragraph')
    if (ast[0].type !== 'paragraph') return
    const kinds = ast[0].children.map((n) => n.type)
    assert.deepEqual(kinds.filter((t) => t !== 'text'), ['strong', 'em', 'underline'])
  })

  it('parses nested italic inside bold', () => {
    const ast = parseCommentMarkdown('**bold *italic* still**')
    assert.equal(ast[0].type, 'paragraph')
    if (ast[0].type !== 'paragraph') return
    const strong = ast[0].children.find((n) => n.type === 'strong')
    assert.equal(strong?.type, 'strong')
    if (strong?.type !== 'strong') return
    assert.equal(strong.children.some((n) => n.type === 'em'), true)
  })

  it('parses unordered and ordered lists', () => {
    const ul = parseCommentMarkdown('- alpha\n- beta')
    assert.equal(ul[0].type, 'bullet_list')
    if (ul[0].type === 'bullet_list') {
      assert.equal(ul[0].children.length, 2)
      assert.equal(ul[0].children[0].type, 'list_item')
    }
    const ol = parseCommentMarkdown('1. first\n2. second')
    assert.equal(ol[0].type, 'ordered_list')
    if (ol[0].type === 'ordered_list') {
      assert.equal(ol[0].children.length, 2)
    }
  })

  it('parses safe links and unwraps unsafe ones', () => {
    const link = parseCommentMarkdown('[docs](https://example.com/a)')[0]
    assert.equal(link.type, 'paragraph')
    if (link.type === 'paragraph') {
      assert.equal(link.children[0].type, 'link')
      if (link.children[0].type === 'link') assert.equal(link.children[0].href, 'https://example.com/a')
    }
    const unsafe = parseCommentMarkdown('[click](javascript:alert(1))')
    const flat = flattenMdNodes(unsafe)
    assert.equal(flat.some((n) => n.type === 'link'), false)
    assert.equal(flat.some((n) => n.type === 'text'), true)
  })
})
