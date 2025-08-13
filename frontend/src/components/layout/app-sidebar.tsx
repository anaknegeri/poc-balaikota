import * as React from 'react'
import { Link, useLocation } from '@tanstack/react-router'
import { AlertTriangle, Camera, Home, Shield } from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'

// Navigation menu items
const navigationItems = [
  {
    title: 'Home',
    url: '/',
    icon: Home,
  },
  {
    title: 'Camera',
    url: '/cameras',
    icon: Camera,
  },
  {
    title: 'Alert',
    url: '/alerts',
    icon: AlertTriangle,
  },
]

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const location = useLocation()

  return (
    <Sidebar collapsible='none' className='w-16 border-r sm:w-20' {...props}>
      <SidebarHeader className='items-center py-4'>
        <Link className='inline-flex' to='/' aria-label='Go to homepage'>
          <div className='rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 p-2 shadow-lg'>
            <Shield className='h-5 w-5 text-white' />
          </div>
        </Link>
      </SidebarHeader>
      <SidebarContent className='overflow-visible'>
        <SidebarGroup className='p-4'>
          <SidebarMenu className='gap-4'>
            {navigationItems.map((item) => {
              const isActive = location.pathname === item.url

              return (
                <SidebarMenuItem
                  key={item.title}
                  className='flex items-center justify-center'
                >
                  <span className='has-data-[active=true]:before:bg-sidebar-primary/48 has-data-[active=true]:after:bg-foreground relative has-data-[active=true]:before:absolute has-data-[active=true]:before:inset-0 has-data-[active=true]:before:-left-2 has-data-[active=true]:before:rounded-full has-data-[active=true]:before:blur-[10px] has-data-[active=true]:after:absolute has-data-[active=true]:after:top-1/2 has-data-[active=true]:after:right-full has-data-[active=true]:after:size-1 has-data-[active=true]:after:-translate-x-2 has-data-[active=true]:after:-translate-y-1/2 has-data-[active=true]:after:rounded-full'>
                    <SidebarMenuButton
                      asChild
                      className='from-background/64 to-background dark:bg-card/64 dark:hover:bg-card/80 relative size-9 items-center justify-center rounded-full bg-linear-to-b p-0 shadow-[0_1px_1px_rgba(0,0,0,0.05),_0_2px_2px_rgba(0,0,0,0.05),_0_4px_4px_rgba(0,0,0,0.05),_0_6px_6px_rgba(0,0,0,0.05)] sm:size-11 dark:bg-none dark:inset-shadow-[0_1px_rgb(255_255_255/0.15)]'
                      tooltip={{
                        children: item.title,
                        hidden: false,
                      }}
                      isActive={isActive}
                    >
                      <Link to={item.url}>
                        <item.icon className='size-5' />
                        <span className='sr-only'>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                  </span>
                </SidebarMenuItem>
              )
            })}
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  )
}
