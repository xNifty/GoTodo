import type { Tag } from '@/api/types'

/** Keep the first tag for each case-insensitive name. */
export function uniqueTagsByName(tags: Tag[]): Tag[] {
  const seen = new Set<string>()
  const out: Tag[] = []
  for (const tag of tags) {
    const key = tag.name.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(tag)
  }
  return out
}
