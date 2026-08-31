import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import {
  altFromFilename,
  armTaskOverlayFileGuard,
  isAllowedImageFile,
  shouldIgnoreTaskOverlayClose,
  toImageMarkdown,
} from './imageUpload.ts'

describe('image markdown', () => {
  it('builds GitHub-style image tags', () => {
    assert.equal(toImageMarkdown('https://cdn.example.com/a.png', 'cat'), '![cat](https://cdn.example.com/a.png)')
    assert.equal(altFromFilename('my (pic).png'), 'my pic')
  })
})

describe('isAllowedImageFile', () => {
  it('accepts png and jpeg', () => {
    assert.equal(isAllowedImageFile(new File([], 'x.png', { type: 'image/png' })), true)
    assert.equal(isAllowedImageFile(new File([], 'x.jpg', { type: 'image/jpeg' })), true)
    assert.equal(isAllowedImageFile(new File([], 'x.txt', { type: 'text/plain' })), false)
  })
})

describe('task overlay file guard', () => {
  it('ignores the click that follows a native file picker', () => {
    armTaskOverlayFileGuard()
    assert.equal(shouldIgnoreTaskOverlayClose(), true)
  })
})
