import { useWebSocketContext } from '@/contexts/websocket-context'
import { Wifi, WifiOff } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

export function WebSocketStatus() {
  const { isConnected, connect } = useWebSocketContext()

  if (isConnected) {
    return (
      <Badge
        variant='outline'
        className='border-green-200 bg-green-50 text-green-700'
      >
        <Wifi className='mr-1 h-3 w-3' />
        Connected
      </Badge>
    )
  }

  return (
    <div className='flex items-center gap-2'>
      <Badge
        variant='outline'
        className='border-red-200 bg-red-50 text-red-700'
      >
        <WifiOff className='mr-1 h-3 w-3' />
        Disconnected
      </Badge>
      <Button
        variant='outline'
        size='sm'
        onClick={connect}
        className='h-6 px-2 text-xs'
      >
        Retry
      </Button>
    </div>
  )
}
