import { useState } from 'react'
import { Outlet, useNavigate } from '@tanstack/react-router'
import { useWebSocketContext } from '@/contexts/websocket-context'
import { AlertTriangle, Bell, Clock } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/layout/app-sidebar'

function NotificationIcon({ severity }: { severity: string }) {
  const getIconColor = () => {
    switch (severity) {
      case 'high':
        return 'text-red-500'
      case 'medium':
        return 'text-amber-500'
      case 'low':
        return 'text-emerald-500'
      default:
        return 'text-gray-500'
    }
  }

  return (
    <div className={`rounded-full p-2 ${getIconColor()}`}>
      <AlertTriangle className='h-4 w-4' />
    </div>
  )
}

export function BaseLayout() {
  const navigate = useNavigate()
  const [selectedImage, setSelectedImage] = useState<string | null>(null)
  const { recentAlerts } = useWebSocketContext()

  // Convert recentAlerts to notification format with image support
  const alertNotifications = recentAlerts.map((alert, index) => ({
    id: alert.id || `alert-${index}`,
    type: alert.data?.alert_type || alert.alert_type || 'security',
    title: `${alert.data?.data?.alert_type?.display_name || alert.data?.alert_type || alert.alert_type || 'Security'} Alert`,
    description: `${alert.data?.message || alert.message} - ${alert.data?.camera_name || alert.camera_name}`,
    time: alert.timestamp
      ? new Date(alert.timestamp).toLocaleString()
      : 'Just now',
    severity: alert.data?.severity || alert.severity || 'medium',
    icon: AlertTriangle,
    unread: true,
    image_url: alert.data?.image_url || alert.image_url, // Add image support
  }))

  // Use only real-time alerts from WebSocket
  const allNotifications = alertNotifications
  const unreadCount = alertNotifications.length

  // Image Modal Component
  const ImageModal = ({
    image,
    onClose,
  }: {
    image: string
    onClose: () => void
  }) => (
    <div
      className='fixed inset-0 z-50 flex items-center justify-center bg-black/80'
      onClick={onClose}
    >
      <div className='relative max-h-[90vh] max-w-4xl p-4'>
        <button
          onClick={onClose}
          className='absolute top-2 right-2 flex h-8 w-8 cursor-pointer items-center justify-center rounded-full bg-white/20 text-xl font-bold text-white transition-colors hover:bg-white/30'
        >
          ×
        </button>
        <img
          src={image}
          alt='Alert snapshot'
          className='max-h-full max-w-full rounded-lg'
        />
      </div>
    </div>
  )

  return (
    <>
      {/* Image Modal */}
      {selectedImage && (
        <ImageModal
          image={selectedImage}
          onClose={() => setSelectedImage(null)}
        />
      )}

      <div className='flex h-svh bg-gradient-to-br from-gray-50 via-white to-gray-100/50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900'>
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset className='overflow-x-hidden overflow-y-auto'>
            {/* Enhanced Header */}
            <header className='sticky top-0 z-50 border-b border-gray-200/80 bg-white/80 backdrop-blur-xl supports-[backdrop-filter]:bg-white/60 dark:border-gray-700/80 dark:bg-gray-800/80 dark:supports-[backdrop-filter]:bg-gray-800/60'>
              <div className='container flex h-16 max-w-screen-2xl items-center justify-between px-6'>
                {/* Left side - Government branding */}
                <div className='flex items-center gap-4'></div>

                {/* Center - Current time */}
                <div className='hidden items-center gap-2 rounded-full bg-gray-100/80 px-4 py-2 md:flex dark:bg-gray-700/50'>
                  <Clock className='h-4 w-4 text-gray-600 dark:text-gray-400' />
                  <span className='text-sm font-medium text-gray-700 dark:text-gray-300'>
                    {new Date().toLocaleTimeString('en-US', {
                      hour: '2-digit',
                      minute: '2-digit',
                      hour12: true,
                    })}
                  </span>
                </div>

                {/* Right side - Actions */}
                <div className='flex items-center gap-3'>
                  {/* Notifications Dropdown */}
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant='ghost'
                        size='sm'
                        className='relative h-10 w-10 rounded-full transition-all duration-200 hover:bg-blue-50 hover:ring-2 hover:ring-blue-200 dark:hover:bg-blue-900/20 dark:hover:ring-blue-800'
                      >
                        <Bell className='h-5 w-5 text-gray-600 dark:text-gray-400' />
                        {unreadCount > 0 && (
                          <Badge className='absolute -top-1 -right-1 h-5 w-5 rounded-full border-2 border-white bg-red-500 p-0 text-[10px] font-bold text-white shadow-lg dark:border-gray-800'>
                            {unreadCount}
                          </Badge>
                        )}
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                      align='end'
                      className='w-96 border-gray-200/50 p-0 shadow-xl dark:border-gray-700/50'
                    >
                      {/* Header */}
                      <div className='border-b border-gray-100 bg-gradient-to-r from-blue-50 to-blue-100/50 p-4 dark:border-gray-700 dark:from-blue-900/20 dark:to-blue-800/20'>
                        <div className='flex items-center justify-between'>
                          <h4 className='text-lg font-bold text-gray-900 dark:text-white'>
                            Security Notifications
                          </h4>
                          {unreadCount > 0 && (
                            <Badge
                              variant='secondary'
                              className='bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300'
                            >
                              {unreadCount} new
                            </Badge>
                          )}
                        </div>
                        <p className='mt-1 text-sm text-gray-600 dark:text-gray-400'>
                          Real-time security alerts and updates
                        </p>
                      </div>

                      {/* Notifications List */}
                      <div className='max-h-96 overflow-y-auto'>
                        <div className='p-2'>
                          {allNotifications.map((notification) => (
                            <div
                              key={notification.id}
                              className={`group cursor-pointer rounded-xl p-4 transition-all duration-200 hover:bg-gray-50 dark:hover:bg-gray-800/50 ${
                                notification.unread
                                  ? 'bg-blue-50/50 dark:bg-blue-900/10'
                                  : ''
                              }`}
                            >
                              <div className='flex items-start gap-4'>
                                {/* Alert Thumbnail */}
                                {notification.image_url && (
                                  <div className='h-12 w-16 flex-shrink-0 overflow-hidden rounded-md bg-gray-100 dark:bg-gray-800'>
                                    <img
                                      src={notification.image_url}
                                      alt={`Alert ${notification.id}`}
                                      className='h-full w-full cursor-pointer object-cover transition-transform hover:scale-105'
                                      onClick={(e) => {
                                        e.stopPropagation()
                                        setSelectedImage(
                                          notification.image_url!
                                        )
                                      }}
                                      onError={(e) => {
                                        const target =
                                          e.target as HTMLImageElement
                                        target.style.display = 'none'
                                        const parent = target.parentElement
                                        if (parent) {
                                          parent.innerHTML = `
                                            <div class="flex h-full w-full items-center justify-center bg-gray-200 dark:bg-gray-700">
                                              <svg class="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                              </svg>
                                            </div>
                                          `
                                        }
                                      }}
                                    />
                                  </div>
                                )}

                                <NotificationIcon
                                  severity={notification.severity}
                                />
                                <div className='min-w-0 flex-1 space-y-2'>
                                  <div className='flex items-start justify-between gap-2'>
                                    <p
                                      className={`text-sm font-semibold ${
                                        notification.unread
                                          ? 'text-gray-900 dark:text-white'
                                          : 'text-gray-700 dark:text-gray-300'
                                      }`}
                                    >
                                      {notification.title}
                                    </p>
                                    {notification.unread && (
                                      <div className='mt-1 h-2 w-2 flex-shrink-0 rounded-full bg-blue-500'></div>
                                    )}
                                  </div>
                                  <p className='text-xs leading-relaxed text-gray-600 dark:text-gray-400'>
                                    {notification.description}
                                  </p>
                                  <div className='flex items-center gap-2'>
                                    <Clock className='h-3 w-3 text-gray-400' />
                                    <p className='text-xs font-medium text-gray-500 dark:text-gray-500'>
                                      {notification.time}
                                    </p>
                                  </div>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>

                      {/* Footer with Navigation */}
                      <div className='border-t border-gray-100 p-3 dark:border-gray-700'>
                        <Button
                          variant='ghost'
                          className='h-9 w-full justify-center text-sm font-medium text-blue-600 hover:bg-blue-50 hover:text-blue-700 dark:text-blue-400 dark:hover:bg-blue-900/20 dark:hover:text-blue-300'
                          onClick={() => navigate({ to: '/alerts' })}
                        >
                          View All Notifications
                        </Button>
                      </div>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            </header>

            {/* Main Content Area */}
            <div className='relative min-h-[calc(100vh-4rem)]'>
              {/* Background Pattern */}
              <div className='pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_1px_1px,rgba(59,130,246,0.1)_1px,transparent_0)] [background-size:24px_24px]' />

              {/* Content */}
              <div className='relative'>
                <Outlet />
              </div>
            </div>
          </SidebarInset>
        </SidebarProvider>
      </div>
    </>
  )
}
