import { useEffect, useRef, useState } from 'react'
import {
  BellIcon,
  BookOpenIcon,
  ChevronDownIcon,
  InfoIcon,
  LifeBuoyIcon,
  LogOutIcon,
  MonitorIcon,
  MoonIcon,
  SettingsIcon,
  SunIcon,
  UserIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTheme } from '@/context/theme-context'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '@/components/ui/navigation-menu'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

// Navigation links array to be used in both desktop and mobile menus
const navigationLinks = [
  { href: '#', label: 'Home' },
  {
    label: 'Features',
    submenu: true,
    type: 'description',
    items: [
      {
        href: '#',
        label: 'Components',
        description: 'Browse all components in the library.',
      },
      {
        href: '#',
        label: 'Documentation',
        description: 'Learn how to use the library.',
      },
      {
        href: '#',
        label: 'Templates',
        description: 'Pre-built layouts for common use cases.',
      },
    ],
  },
  {
    label: 'Pricing',
    submenu: true,
    type: 'simple',
    items: [
      { href: '#', label: 'Product A' },
      { href: '#', label: 'Product B' },
      { href: '#', label: 'Product C' },
      { href: '#', label: 'Product D' },
    ],
  },
  {
    label: 'About',
    submenu: true,
    type: 'icon',
    items: [
      { href: '#', label: 'Getting Started', icon: 'BookOpenIcon' },
      { href: '#', label: 'Tutorials', icon: 'LifeBuoyIcon' },
      { href: '#', label: 'About Us', icon: 'InfoIcon' },
    ],
  },
]

export default function AppNavbar() {
  const { theme, setTheme } = useTheme()
  const [scrolled, setScrolled] = useState(false)
  const [hidden, setHidden] = useState(false)
  const lastScrollYRef = useRef(0)
  const hideTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    const handleScroll = () => {
      const currentScrollY = window.scrollY
      const isScrollingDown = currentScrollY > lastScrollYRef.current
      const scrollThreshold = 100 // Minimum scroll distance to trigger hide

      // Clear existing timeout
      if (hideTimeoutRef.current) {
        clearTimeout(hideTimeoutRef.current)
        hideTimeoutRef.current = null
      }

      // Only hide when scrolling down from top and past threshold
      if (isScrollingDown && currentScrollY > scrollThreshold) {
        setHidden(true)

        // Show navbar again after 1.5 seconds
        hideTimeoutRef.current = setTimeout(() => {
          setHidden(false)
          hideTimeoutRef.current = null
        }, 300)
      } else if (!isScrollingDown || currentScrollY <= scrollThreshold) {
        // Show immediately when scrolling up or near top
        setHidden(false)
      }

      // Set scrolled state for styling
      setScrolled(currentScrollY > 10)
      lastScrollYRef.current = currentScrollY
    }

    window.addEventListener('scroll', handleScroll, { passive: true })

    return () => {
      window.removeEventListener('scroll', handleScroll)
      if (hideTimeoutRef.current) clearTimeout(hideTimeoutRef.current)
    }
  }, [])

  return (
    <header
      className={cn(
        'fixed top-0 right-0 z-50 transition-all duration-300 ease-in-out',
        'border-b px-4 backdrop-blur-lg md:px-6',
        'left-0 md:left-[15.6rem]', // Offset untuk sidebar di desktop
        scrolled
          ? 'bg-background/95 border-border shadow-md'
          : 'bg-background/90 border-border/60',
        hidden ? '-translate-y-full' : 'translate-y-0'
      )}
    >
      <div className='flex h-16 items-center justify-between gap-4'>
        {/* Left side */}
        <div className='flex items-center gap-2'>
          {/* Mobile menu trigger */}
          <Popover>
            <PopoverTrigger asChild>
              <Button
                className='group size-8 md:hidden'
                variant='ghost'
                size='icon'
              >
                <svg
                  className='pointer-events-none'
                  width={16}
                  height={16}
                  viewBox='0 0 24 24'
                  fill='none'
                  stroke='currentColor'
                  strokeWidth='2'
                  strokeLinecap='round'
                  strokeLinejoin='round'
                  xmlns='http://www.w3.org/2000/svg'
                >
                  <path
                    d='M4 12L20 12'
                    className='origin-center -translate-y-[7px] transition-all duration-300 ease-[cubic-bezier(.5,.85,.25,1.1)] group-aria-expanded:translate-x-0 group-aria-expanded:translate-y-0 group-aria-expanded:rotate-[315deg]'
                  />
                  <path
                    d='M4 12H20'
                    className='origin-center transition-all duration-300 ease-[cubic-bezier(.5,.85,.25,1.8)] group-aria-expanded:rotate-45'
                  />
                  <path
                    d='M4 12H20'
                    className='origin-center translate-y-[7px] transition-all duration-300 ease-[cubic-bezier(.5,.85,.25,1.1)] group-aria-expanded:translate-y-0 group-aria-expanded:rotate-[135deg]'
                  />
                </svg>
              </Button>
            </PopoverTrigger>
            <PopoverContent align='start' className='w-64 p-1 md:hidden'>
              <NavigationMenu className='max-w-none *:w-full'>
                <NavigationMenuList className='flex-col items-start gap-0 md:gap-2'>
                  {navigationLinks.map((link, index) => (
                    <NavigationMenuItem key={index} className='w-full'>
                      {link.submenu ? (
                        <>
                          <div className='text-muted-foreground px-2 py-1.5 text-xs font-medium'>
                            {link.label}
                          </div>
                          <ul>
                            {link.items.map((item, itemIndex) => (
                              <li key={itemIndex}>
                                <NavigationMenuLink
                                  href={item.href}
                                  className='py-1.5'
                                >
                                  {item.label}
                                </NavigationMenuLink>
                              </li>
                            ))}
                          </ul>
                        </>
                      ) : (
                        <NavigationMenuLink href={link.href} className='py-1.5'>
                          {link.label}
                        </NavigationMenuLink>
                      )}
                      {/* Add separator between different types of items */}
                      {index < navigationLinks.length - 1 &&
                        // Show separator if:
                        // 1. One is submenu and one is simple link OR
                        // 2. Both are submenus but with different types
                        ((!link.submenu &&
                          navigationLinks[index + 1].submenu) ||
                          (link.submenu &&
                            !navigationLinks[index + 1].submenu) ||
                          (link.submenu &&
                            navigationLinks[index + 1].submenu &&
                            link.type !== navigationLinks[index + 1].type)) && (
                          <div
                            role='separator'
                            aria-orientation='horizontal'
                            className='bg-border -mx-1 my-1 h-px w-full'
                          />
                        )}
                    </NavigationMenuItem>
                  ))}
                </NavigationMenuList>
              </NavigationMenu>
            </PopoverContent>
          </Popover>
          {/* Main nav */}
          <div className='flex items-center gap-6'>
            {/* Navigation menu */}
            <NavigationMenu viewport={false} className='max-md:hidden'>
              <NavigationMenuList className='gap-2'>
                {navigationLinks.map((link, index) => (
                  <NavigationMenuItem key={index}>
                    {link.submenu ? (
                      <>
                        <NavigationMenuTrigger className='text-muted-foreground hover:text-primary bg-transparent px-2 py-1.5 font-medium *:[svg]:-me-0.5 *:[svg]:size-3.5'>
                          {link.label}
                        </NavigationMenuTrigger>
                        <NavigationMenuContent className='data-[motion=from-end]:slide-in-from-right-16! data-[motion=from-start]:slide-in-from-left-16! data-[motion=to-end]:slide-out-to-right-16! data-[motion=to-start]:slide-out-to-left-16! z-50 p-1'>
                          <ul
                            className={cn(
                              link.type === 'description'
                                ? 'min-w-64'
                                : 'min-w-48'
                            )}
                          >
                            {link.items.map((item, itemIndex) => (
                              <li key={itemIndex}>
                                <NavigationMenuLink
                                  href={item.href}
                                  className='py-1.5'
                                >
                                  {/* Display icon if present */}
                                  {link.type === 'icon' && 'icon' in item && (
                                    <div className='flex items-center gap-2'>
                                      {item.icon === 'BookOpenIcon' && (
                                        <BookOpenIcon
                                          size={16}
                                          className='text-foreground opacity-60'
                                          aria-hidden='true'
                                        />
                                      )}
                                      {item.icon === 'LifeBuoyIcon' && (
                                        <LifeBuoyIcon
                                          size={16}
                                          className='text-foreground opacity-60'
                                          aria-hidden='true'
                                        />
                                      )}
                                      {item.icon === 'InfoIcon' && (
                                        <InfoIcon
                                          size={16}
                                          className='text-foreground opacity-60'
                                          aria-hidden='true'
                                        />
                                      )}
                                      <span>{item.label}</span>
                                    </div>
                                  )}

                                  {/* Display label with description if present */}
                                  {link.type === 'description' &&
                                  'description' in item ? (
                                    <div className='space-y-1'>
                                      <div className='font-medium'>
                                        {item.label}
                                      </div>
                                      <p className='text-muted-foreground line-clamp-2 text-xs'>
                                        {item.description}
                                      </p>
                                    </div>
                                  ) : (
                                    // Display simple label if not icon or description type
                                    !link.type ||
                                    (link.type !== 'icon' &&
                                      link.type !== 'description' && (
                                        <span>{item.label}</span>
                                      ))
                                  )}
                                </NavigationMenuLink>
                              </li>
                            ))}
                          </ul>
                        </NavigationMenuContent>
                      </>
                    ) : (
                      <NavigationMenuLink
                        href={link.href}
                        className='text-muted-foreground hover:text-primary py-1.5 font-medium'
                      >
                        {link.label}
                      </NavigationMenuLink>
                    )}
                  </NavigationMenuItem>
                ))}
              </NavigationMenuList>
            </NavigationMenu>
          </div>
        </div>
        {/* User menu */}
        <div className='flex items-center gap-2'>
          {/* Notification button */}
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='hover:bg-muted/80 relative transition-all duration-200 hover:scale-105'
              >
                <BellIcon size={18} />
                <span className='absolute -top-1 -right-1 size-2 animate-pulse rounded-full bg-red-500'></span>
              </Button>
            </PopoverTrigger>
            <PopoverContent align='end' className='w-72 p-1'>
              <div className='space-y-1'>
                <div className='px-3 py-2 text-sm font-medium'>
                  Notifications
                </div>
                <div className='space-y-1'>
                  <div className='hover:bg-muted cursor-pointer rounded-sm px-3 py-2 text-sm'>
                    <div className='font-medium'>New message received</div>
                    <div className='text-muted-foreground text-xs'>
                      2 minutes ago
                    </div>
                  </div>
                  <div className='hover:bg-muted cursor-pointer rounded-sm px-3 py-2 text-sm'>
                    <div className='font-medium'>System update available</div>
                    <div className='text-muted-foreground text-xs'>
                      1 hour ago
                    </div>
                  </div>
                  <div className='hover:bg-muted cursor-pointer rounded-sm px-3 py-2 text-sm'>
                    <div className='font-medium'>Welcome to the platform!</div>
                    <div className='text-muted-foreground text-xs'>
                      1 day ago
                    </div>
                  </div>
                </div>
              </div>
            </PopoverContent>
          </Popover>

          {/* Theme switcher */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                className='hover:bg-muted/80 transition-all duration-200 hover:scale-105 hover:rotate-12'
              >
                {theme === 'light' && <SunIcon size={18} />}
                {theme === 'dark' && <MoonIcon size={18} />}
                {theme === 'system' && <MonitorIcon size={18} />}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align='end'>
              <DropdownMenuItem onClick={() => setTheme('light')}>
                <SunIcon size={16} className='mr-2' />
                Light
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme('dark')}>
                <MoonIcon size={16} className='mr-2' />
                Dark
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme('system')}>
                <MonitorIcon size={16} className='mr-2' />
                System
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          {/* User menu */}
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant='ghost'
                size='sm'
                className='hover:bg-muted/80 flex items-center gap-2 text-sm transition-all duration-200 hover:scale-105'
              >
                <div className='bg-primary flex size-6 items-center justify-center rounded-full'>
                  <UserIcon size={14} className='text-primary-foreground' />
                </div>
                <span>User</span>
                <ChevronDownIcon size={14} className='opacity-50' />
              </Button>
            </PopoverTrigger>
            <PopoverContent align='end' className='w-48 p-1'>
              <div className='space-y-1'>
                <Button
                  variant='ghost'
                  size='sm'
                  className='w-full justify-start text-sm'
                >
                  <UserIcon size={16} className='mr-2' />
                  Profile
                </Button>
                <Button
                  variant='ghost'
                  size='sm'
                  className='w-full justify-start text-sm'
                >
                  <SettingsIcon size={16} className='mr-2' />
                  Settings
                </Button>
                <div className='my-1 border-t' />
                <Button
                  variant='ghost'
                  size='sm'
                  className='w-full justify-start text-sm'
                >
                  <LogOutIcon size={16} className='mr-2' />
                  Sign Out
                </Button>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>
    </header>
  )
}
