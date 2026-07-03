import { useMemo, useState } from 'react'
import { SearchIcon } from 'lucide-react'

import { Input } from '@/components/ui/input'
import { useProjects } from './use-projects'
import { ProjectCard } from './project-card'

/**
 * The landing dashboard: the project list plus a live client-side filter.
 * The list endpoint has no name-search param, so filtering happens in memory
 * over the already-fetched list (no refetch on keystroke).
 */
export function ProjectList() {
  const { data: projects, isLoading, isError, error } = useProjects()
  const [query, setQuery] = useState('')

  const filtered = useMemo(() => {
    const list = projects ?? []
    const q = query.trim().toLowerCase()
    if (!q) return list
    return list.filter((p) => p.project_name.toLowerCase().includes(q))
  }, [projects, query])

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8">
      <div className="mb-6 flex flex-col gap-1">
        <h1 className="text-2xl font-semibold tracking-tight">Projects</h1>
        <p className="text-muted-foreground text-sm">
          {projects ? `${projects.length} project(s)` : ' '}
        </p>
      </div>

      <div className="relative mb-6 max-w-sm">
        <SearchIcon className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2" />
        <Input
          type="search"
          placeholder="Filter projects…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
          aria-label="Filter projects by name"
        />
      </div>

      {isLoading && (
        <p className="text-muted-foreground text-sm">Loading projects…</p>
      )}

      {isError && (
        <p className="text-destructive text-sm">
          Failed to load projects{error instanceof Error ? `: ${error.message}` : ''}.
        </p>
      )}

      {!isLoading && !isError && filtered.length === 0 && (
        <p className="text-muted-foreground text-sm">
          {projects && projects.length > 0
            ? 'No projects match your filter.'
            : 'No projects yet.'}
        </p>
      )}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {filtered.map((project) => (
          <ProjectCard key={project.project_identifier} project={project} />
        ))}
      </div>
    </div>
  )
}
