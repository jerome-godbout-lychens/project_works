import { createFileRoute } from '@tanstack/react-router'

import { ProjectList } from '@/features/projects/project-list'

export const Route = createFileRoute('/_authed/')({
  component: ProjectList,
})
