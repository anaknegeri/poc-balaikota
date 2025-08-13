import { createFileRoute } from '@tanstack/react-router'
import Dashboard from '@/features/dashboard'

export const Route = createFileRoute('/(pages)/')({
  component: RouteComponent,
})

function RouteComponent() {
  return <Dashboard />
}
