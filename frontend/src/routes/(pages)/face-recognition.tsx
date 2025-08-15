import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/(pages)/face-recognition')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/(pages)/face-recognition"!</div>
}
