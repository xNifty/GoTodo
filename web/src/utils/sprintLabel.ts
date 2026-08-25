import type { ProjectSprint } from '@/api/types'

/** Label for sprint selects: "3.0.0 - features required for v3.0.0 release". */
export function sprintOptionLabel(sprint: ProjectSprint, opts?: { activeSuffix?: boolean }): string {
  // const desc = (sprint.description || '').trim()
  // const base = desc ? `${sprint.name} - ${desc}` : sprint.name
  const base = sprint.name
  if (opts?.activeSuffix && sprint.is_active) {
    return `${base} (active)`
  }
  return base
}
