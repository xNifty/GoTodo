import type { Project } from '@/api/types'

/** Label for project selects: shared projects include role, e.g. "Shared One (viewer)". */
export function projectOptionLabel(project: Project): string {
  if (project.role && project.role !== 'owner') {
    return `${project.name} (${project.role})`
  }
  return project.name
}
