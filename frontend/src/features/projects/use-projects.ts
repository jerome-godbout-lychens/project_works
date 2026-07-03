import { useQuery } from '@tanstack/react-query'
import { listProjects } from '@/api'
import type { Project } from '@/api'

/** Loads the flat project list from GET /projects. */
export function useProjects() {
  return useQuery<Project[]>({
    queryKey: ['projects'],
    queryFn: async () => {
      const { data } = await listProjects({ throwOnError: true })
      return data.items
    },
  })
}
