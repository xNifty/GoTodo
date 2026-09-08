import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { continueListOnEnter, insertLinkMarkup, normalizeLinkHref, prefixLines, wrapInline, wouldExceedLimit } from './markdownToolbar.ts'

describe('wrapInline', () => {
  it('wraps a selection', () => {
    const got = wrapInline('say hi now', 4, 6, '**')
    assert.equal(got.body, 'say **hi** now')
    assert.equal(got.start, 10)
    assert.equal(got.end, 10)
  })

  it('inserts a placeholder when the selection is empty', () => {
    const got = wrapInline('ab', 1, 1, '*')
    assert.equal(got.body, 'a*text*b')
    assert.equal(got.start, 2)
    assert.equal(got.end, 6)
  })

  it('unwraps when the selection already includes markers', () => {
    const got = wrapInline('say **hi** now', 4, 10, '**')
    assert.equal(got.body, 'say hi now')
  })

  it('unwraps when markers surround the selection', () => {
    const got = wrapInline('say **hi** now', 6, 8, '**')
    assert.equal(got.body, 'say hi now')
    assert.equal(got.start, 4)
    assert.equal(got.end, 6)
  })

  it('wraps underline with plus markers', () => {
    const got = wrapInline('hi', 0, 2, '++')
    assert.equal(got.body, '++hi++')
  })
})

describe('prefixLines', () => {
  it('prefixes an unordered list', () => {
    const got = prefixLines('alpha\nbeta', 0, 10, 'ul')
    assert.equal(got.body, '- alpha\n- beta')
  })

  it('prefixes an ordered list', () => {
    const got = prefixLines('alpha\nbeta', 0, 10, 'ol')
    assert.equal(got.body, '1. alpha\n2. beta')
  })

  it('toggles the same list kind off', () => {
    const got = prefixLines('- alpha\n- beta', 0, 14, 'ul')
    assert.equal(got.body, 'alpha\nbeta')
  })

  it('converts ordered lines to unordered', () => {
    const got = prefixLines('1. alpha\n2. beta', 0, 16, 'ul')
    assert.equal(got.body, '- alpha\n- beta')
  })

  it('inserts a list marker on an empty line', () => {
    const got = prefixLines('', 0, 0, 'ul')
    assert.equal(got.body, '- ')
  })
})

describe('insertLinkMarkup', () => {
  it('uses the title and href and places the cursor after the link', () => {
    const got = insertLinkMarkup('see docs here', 4, 8, 'docs', 'https://example.com/')
    assert.equal(got.body, 'see [docs](https://example.com/) here')
    assert.equal(got.start, 32)
    assert.equal(got.end, 32)
  })

  it('falls back to the URL when the title is empty', () => {
    const got = insertLinkMarkup('', 0, 0, '  ', 'https://example.com/')
    assert.equal(got.body, '[https://example.com/](https://example.com/)')
  })
})

describe('continueListOnEnter', () => {
  it('starts the next unordered item', () => {
    const got = continueListOnEnter('- hello', 7)
    assert.ok(got)
    assert.equal(got?.body, '- hello\n- ')
    assert.equal(got?.start, 10)
  })

  it('splits an unordered item at the cursor', () => {
    const got = continueListOnEnter('- hello', 4)
    assert.equal(got?.body, '- he\n- llo')
  })

  it('increments the next ordered item', () => {
    const got = continueListOnEnter('1. hello', 8)
    assert.equal(got?.body, '1. hello\n2. ')
  })

  it('leaves the list on an empty item', () => {
    const got = continueListOnEnter('- hello\n- ', 10)
    assert.equal(got?.body, '- hello\n')
    assert.equal(got?.start, 8)
  })

  it('does nothing on a normal paragraph', () => {
    assert.equal(continueListOnEnter('hello', 5), null)
  })
})

describe('normalizeLinkHref', () => {
  it('adds https when the scheme is missing', () => {
    assert.equal(normalizeLinkHref('example.com/a'), 'https://example.com/a')
  })

  it('rejects javascript URLs', () => {
    assert.equal(normalizeLinkHref('javascript:alert(1)'), null)
    assert.equal(normalizeLinkHref(''), null)
  })
})

describe('wouldExceedLimit', () => {
  it('detects overflow', () => {
    assert.equal(wouldExceedLimit(2001, 2000), true)
    assert.equal(wouldExceedLimit(2000, 2000), false)
    assert.equal(wouldExceedLimit(10, 0), false)
  })
})
