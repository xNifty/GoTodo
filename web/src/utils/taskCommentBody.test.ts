import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import {
  hasImageMarkdown,
  insertMarkdownAtCursor,
  isSafeImageSrc,
  previewWithoutImages,
  splitCommentBody,
} from './taskCommentBody.ts'

describe('isSafeImageSrc', () => {
  it('allows https and local uploads', () => {
    assert.equal(isSafeImageSrc('https://cdn.example.com/a.png'), true)
    assert.equal(isSafeImageSrc('/uploads/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee.png'), true)
  })
  it('rejects javascript and relative tricks', () => {
    assert.equal(isSafeImageSrc('javascript:alert(1)'), false)
    assert.equal(isSafeImageSrc('data:image/png;base64,aaa'), false)
    assert.equal(isSafeImageSrc('/uploads/../secret.png'), false)
    assert.equal(isSafeImageSrc('//evil.example/x.png'), false)
  })
})

describe('splitCommentBody images', () => {
  it('keeps mentions and task refs working', () => {
    const parts = splitCommentBody('Hi @Ada see #12 and [[9]]')
    const types = parts.map((p) => p.type)
    assert.deepEqual(types, ['text', 'mention', 'text', 'task', 'text', 'task'])
  })

  it('renders inserted markdown as an image part', () => {
    const parts = splitCommentBody('see ![cat](https://cdn.example.com/cat.png) please')
    assert.equal(parts.length, 3)
    assert.equal(parts[0].type, 'text')
    assert.equal(parts[1].type, 'image')
    if (parts[1].type === 'image') {
      assert.equal(parts[1].alt, 'cat')
      assert.equal(parts[1].src, 'https://cdn.example.com/cat.png')
    }
    assert.equal(parts[2].type, 'text')
  })

  it('does not treat unsafe URLs as images', () => {
    const parts = splitCommentBody('![x](javascript:alert(1))')
    assert.equal(parts.some((p) => p.type === 'image'), false)
  })

  it('does not eat a following task number out of a URL', () => {
    const parts = splitCommentBody('![shot](https://cdn.example.com/a.png)\n#42')
    assert.equal(parts[0].type, 'image')
    assert.equal(parts[parts.length - 1].type, 'task')
  })
})

describe('previewWithoutImages', () => {
  it('replaces markdown with a short label', () => {
    assert.equal(
      previewWithoutImages('Before ![logo](https://cdn.example.com/a.png) after'),
      'Before [image: logo] after',
    )
  })
})

describe('hasImageMarkdown', () => {
  it('detects a safe image', () => {
    assert.equal(hasImageMarkdown('![a](https://cdn.example.com/a.png)'), true)
    assert.equal(hasImageMarkdown('no pictures here'), false)
  })
})

describe('insertMarkdownAtCursor', () => {
  it('puts the image on its own line', () => {
    const got = insertMarkdownAtCursor('hello', '![cat](https://cdn.example.com/c.png)', 5)
    assert.equal(got.body, 'hello\n![cat](https://cdn.example.com/c.png)')
  })
})
