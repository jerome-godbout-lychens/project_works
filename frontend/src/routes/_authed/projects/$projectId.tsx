import { createFileRoute } from '@tanstack/react-router'

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export const Route = createFileRoute('/_authed/projects/$projectId')({
  component: ProjectDetailStub,
})

/** Placeholder detail view — the real project view is intentionally not built yet. */
function ProjectDetailStub() {
  const { projectId } = Route.useParams()
  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8">
      <Card className="max-w-lg">
        <CardHeader>
          <CardTitle>Project detail</CardTitle>
          <CardDescription>Not implemented yet.</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">
            Project ID: <code className="text-foreground">{projectId}</code>
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
