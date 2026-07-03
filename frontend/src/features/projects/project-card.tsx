import { Link } from '@tanstack/react-router'
import { FolderIcon } from 'lucide-react'

import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import type { Project } from '@/api'

/** A clickable project tile linking to the (stub) project detail route. */
export function ProjectCard({ project }: { project: Project }) {
  return (
    <Link
      to="/projects/$projectId"
      params={{ projectId: project.project_identifier }}
      className="block rounded-xl outline-none focus-visible:ring-ring/50 focus-visible:ring-[3px]"
    >
      <Card className="h-full gap-3 py-4 transition-colors hover:border-ring hover:bg-accent/40">
        <CardHeader className="px-4">
          <CardTitle className="flex items-center gap-2 text-base">
            <FolderIcon className="text-muted-foreground size-4 shrink-0" />
            <span className="truncate">{project.project_name}</span>
          </CardTitle>
          {project.project_description ? (
            <CardDescription className="line-clamp-2">
              {project.project_description}
            </CardDescription>
          ) : project.folder_path ? (
            <CardDescription className="truncate">
              {project.folder_path}
            </CardDescription>
          ) : null}
        </CardHeader>
      </Card>
    </Link>
  )
}
