import type { Project } from '@/api/types'

export function isArchivedProject(project: Project): boolean {
  return !!project.archived
}

/** Label for project selects: shared projects include role, e.g. "Shared One (viewer)". */
export function projectOptionLabel(project: Project): string {
  const base =
    project.role && project.role !== 'owner' ? `${project.name} (${project.role})` : project.name
  return isArchivedProject(project) ? `${base} (archived)` : base
}

export function activeProjects(projects: Project[]): Project[] {
  return projects.filter((p) => !isArchivedProject(p))
}

export function archivedProjects(projects: Project[]): Project[] {
  return projects.filter((p) => isArchivedProject(p))
}

export function isProjectOwner(project: Project): boolean {
  return !project.role || project.role === 'owner'
}
