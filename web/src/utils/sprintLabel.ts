import type { ProjectSprint } from '@/api/types'

/** Label for sprint selects: name plus optional active/locked markers. */
export function sprintOptionLabel(
  sprint: ProjectSprint,
  opts?: { activeSuffix?: boolean; lockedSuffix?: boolean },
): string {
  const base = sprint.name
  const parts: string[] = []
  if (opts?.activeSuffix && sprint.is_active) parts.push('active')
  if (opts?.lockedSuffix && sprint.is_locked) parts.push('locked')
  if (parts.length) return `${base} (${parts.join(', ')})`
  return base
}

export function sprintLockedForUser(sprint: ProjectSprint, role?: string | null): boolean {
  return !!sprint.is_locked && role !== 'owner'
}
