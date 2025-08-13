import { useState } from 'react'
import {
  RiArrowRightLine,
  RiCheckLine,
  RiCloseLine,
  RiRefreshLine,
  RiArrowLeftSLine,
  RiArrowRightSLine,
} from '@remixicon/react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { buttonVariants } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
} from '@/components/ui/pagination'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface Transaction {
  id: string
  date: string
  in: {
    symbol: string
    name: string
    icon: string[]
  }
  out: {
    symbol: string
    name: string
    icon: string[]
  }
  fees: string
  change: {
    received: string
    spent: string
  }
  status: 'completed' | 'failed'
  spent: string
}

const basePath = 'https://raw.githubusercontent.com/dlzlfasou/image/upload/'

const items: Transaction[] = [
  {
    id: '1',
    date: '17 Feb, 2025',
    in: {
      symbol: 'ARK',
      name: 'ArkFi',
      icon: [
        'v1741861900/coin-01-sm-light_hgzpka.svg',
        'v1741861900/coin-01-sm-dark_hkrvvm.svg',
      ],
    },
    out: {
      symbol: 'TOK',
      name: 'Token',
      icon: [
        'v1741861900/coin-02-sm-light_t1qflr.svg',
        'v1741861900/coin-02-sm-dark_iqldgv.svg',
      ],
    },
    fees: '$31.2',
    change: {
      received: '14,972',
      spent: '7,872.1',
    },
    status: 'completed',
    spent: '$2,867.14',
  },
  {
    id: '2',
    date: '17 Feb, 2025',
    in: {
      symbol: 'ARK',
      name: 'ArkFi',
      icon: [
        'v1741861900/coin-01-sm-light_hgzpka.svg',
        'v1741861900/coin-01-sm-dark_hkrvvm.svg',
      ],
    },
    out: {
      symbol: 'TOK',
      name: 'Token',
      icon: [
        'v1741861900/coin-02-sm-light_t1qflr.svg',
        'v1741861900/coin-02-sm-dark_iqldgv.svg',
      ],
    },
    fees: '$22.3',
    change: {
      received: '19,883',
      spent: '12,327',
    },
    status: 'completed',
    spent: '$21,314.24',
  },
  {
    id: '3',
    date: '17 Feb, 2025',
    in: {
      symbol: 'ARK',
      name: 'ArkFi',
      icon: [
        'v1741861900/coin-01-sm-light_hgzpka.svg',
        'v1741861900/coin-01-sm-dark_hkrvvm.svg',
      ],
    },
    out: {
      symbol: 'TOK',
      name: 'Token',
      icon: [
        'v1741861900/coin-02-sm-light_t1qflr.svg',
        'v1741861900/coin-02-sm-dark_iqldgv.svg',
      ],
    },
    fees: '$7.4',
    change: {
      received: '12,487',
      spent: '4,494.2',
    },
    status: 'completed',
    spent: '$1,429.1',
  },
  {
    id: '4',
    date: '17 Feb, 2025',
    in: {
      symbol: 'ARK',
      name: 'ArkFi',
      icon: [
        'v1741861900/coin-01-sm-light_hgzpka.svg',
        'v1741861900/coin-01-sm-dark_hkrvvm.svg',
      ],
    },
    out: {
      symbol: 'TOK',
      name: 'Token',
      icon: [
        'v1741861900/coin-02-sm-light_t1qflr.svg',
        'v1741861900/coin-02-sm-dark_iqldgv.svg',
      ],
    },
    fees: '$42.1',
    change: {
      received: '13,229',
      spent: '7,872.1',
    },
    status: 'completed',
    spent: '$3,411.21',
  },
  {
    id: '5',
    date: '17 Feb, 2025',
    in: {
      symbol: 'ARK',
      name: 'ArkFi',
      icon: [
        'v1741861900/coin-01-sm-light_hgzpka.svg',
        'v1741861900/coin-01-sm-dark_hkrvvm.svg',
      ],
    },
    out: {
      symbol: 'TOK',
      name: 'Token',
      icon: [
        'v1741861900/coin-02-sm-light_t1qflr.svg',
        'v1741861900/coin-02-sm-dark_iqldgv.svg',
      ],
    },
    fees: '$24.7',
    change: {
      received: '14,457',
      spent: '12,226',
    },
    status: 'completed',
    spent: '$12,317.9',
  },
]

export function TransactionsTable() {
  const [currentPage] = useState(1)
  const totalPages = 12

  return (
    <Card className='gap-4'>
      <CardHeader>
        <CardTitle>Transactions</CardTitle>
      </CardHeader>
      <CardContent className='px-0'>
        <Table className='[&_tr_td]:border-border/64 dark:[&_tr_td]:border-card/80 border-separate border-spacing-0 px-6 [&_tr_td]:border-b'>
          <TableHeader>
            <TableRow className='border-0 hover:bg-transparent'>
              <TableHead className='bg-muted dark:bg-card/48 relative h-8 border-0 font-normal select-none first:rounded-l-lg last:rounded-r-lg'>
                Date
              </TableHead>
              <TableHead className='bg-muted dark:bg-card/48 relative h-8 border-0 font-normal select-none first:rounded-l-lg last:rounded-r-lg'>
                Conversion
              </TableHead>
              <TableHead className='bg-muted dark:bg-card/48 relative h-8 border-0 font-normal select-none first:rounded-l-lg last:rounded-r-lg'>
                Fees
              </TableHead>
              <TableHead className='bg-muted dark:bg-card/48 relative h-8 border-0 font-normal select-none first:rounded-l-lg last:rounded-r-lg'>
                Change
              </TableHead>
              <TableHead className='bg-muted dark:bg-card/48 relative h-8 border-0 text-center font-normal select-none first:rounded-l-lg last:rounded-r-lg'>
                Status
              </TableHead>
              <TableHead className='bg-muted dark:bg-card/48 relative h-8 border-0 font-normal select-none first:rounded-l-lg last:rounded-r-lg'>
                Spent
              </TableHead>
              <TableHead className='bg-muted dark:bg-card/48 relative h-8 border-0 text-center font-normal select-none first:rounded-l-lg last:rounded-r-lg'>
                Action
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.id} className='hover:bg-transparent'>
                <TableCell className='text-foreground/70 whitespace-nowrap'>
                  {item.date}
                </TableCell>
                <TableCell className='font-medium'>
                  <div className='flex items-center gap-3'>
                    <div className='flex items-center gap-2'>
                      <div className='shrink-0 rounded-full shadow-xs max-[1320px]:hidden'>
                        <img
                          className='dark:hidden'
                          src={basePath + item.in.icon[0]}
                          alt={item.in.name}
                        />
                        <img
                          className='hidden dark:block'
                          src={basePath + item.in.icon[1]}
                          alt={item.in.name}
                        />
                      </div>
                      <span>{item.in.symbol}</span>
                    </div>
                    <RiArrowRightLine
                      size={16}
                      className='text-muted-foreground/50'
                      aria-hidden='true'
                    />
                    <div className='flex items-center gap-2'>
                      <div className='shrink-0 rounded-full shadow-xs max-[1320px]:hidden'>
                        <img
                          className='dark:hidden'
                          src={basePath + item.out.icon[0]}
                          alt={item.out.name}
                        />
                        <img
                          className='hidden dark:block'
                          src={basePath + item.out.icon[1]}
                          alt={item.out.name}
                        />
                      </div>
                      <span>{item.out.symbol}</span>
                    </div>
                  </div>
                </TableCell>
                <TableCell className='text-foreground/70'>
                  {item.fees}
                </TableCell>
                <TableCell>
                  <div className='flex items-center gap-3'>
                    <Badge className='border-0 bg-emerald-500/12 px-2 py-0.5 text-sm font-normal text-emerald-600'>
                      {item.change.received}
                    </Badge>
                    <RiArrowRightLine
                      size={16}
                      className='text-muted-foreground/50'
                      aria-hidden='true'
                    />
                    <Badge className='border-0 bg-red-500/12 px-2 py-0.5 text-sm font-normal text-red-500'>
                      {item.change.spent}
                    </Badge>
                  </div>
                </TableCell>
                <TableCell className='text-center'>
                  {item.status === 'completed' && (
                    <>
                      <span className='sr-only'>Completed</span>
                      <RiCheckLine
                        size={16}
                        className='inline-flex text-emerald-500'
                      />
                    </>
                  )}
                  {item.status === 'failed' && (
                    <>
                      <span className='sr-only'>Failed</span>
                      <RiCloseLine
                        size={16}
                        className='inline-flex text-red-500'
                      />
                    </>
                  )}
                </TableCell>
                <TableCell className='font-medium'>{item.spent}</TableCell>
                <TableCell className='py-0 text-center'>
                  <Button
                    size='icon'
                    variant='ghost'
                    className='text-muted-foreground/50 dark:hover:bg-card/64 shadow-none'
                    aria-label='Edit item'
                  >
                    <RiRefreshLine size={16} aria-hidden='true' />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <Pagination className='mt-5 px-6'>
          <PaginationContent className='w-full justify-between'>
            <PaginationItem>
              <PaginationLink
                className={cn(
                  buttonVariants({
                    variant: 'outline',
                  }),
                  'aria-disabled:text-muted-foreground/50 hover:bg-muted/50 dark:bg-card/64 dark:hover:bg-card/80 size-8 border-none p-0 shadow-[0px_0px_0px_1px_rgba(0,0,0,0.04),0_1px_1px_rgba(0,0,0,0.05),0_2px_2px_rgba(0,0,0,0.05),0_2px_4px_rgba(0,0,0,0.05)] aria-disabled:pointer-events-none dark:inset-shadow-[0_1px_rgb(255_255_255/0.15)]'
                )}
                href={
                  currentPage === 1 ? undefined : `#/page/${currentPage - 1}`
                }
                aria-label='Go to previous page'
                aria-disabled={currentPage === 1 ? true : undefined}
                role={currentPage === 1 ? 'link' : undefined}
              >
                <RiArrowLeftSLine
                  className='size-5'
                  size={20}
                  aria-hidden='true'
                />
              </PaginationLink>
            </PaginationItem>
            <PaginationItem>
              <p className='text-muted-foreground text-sm' aria-live='polite'>
                Page <span className='text-foreground'>{currentPage}</span> of{' '}
                <span className='text-foreground'>{totalPages}</span>
              </p>
            </PaginationItem>
            <PaginationItem>
              <PaginationLink
                className={cn(
                  buttonVariants({
                    variant: 'outline',
                  }),
                  'aria-disabled:text-muted-foreground/50 hover:bg-muted/50 dark:bg-card/64 dark:hover:bg-card/80 size-8 border-none p-0 shadow-[0px_0px_0px_1px_rgba(0,0,0,0.04),0_1px_1px_rgba(0,0,0,0.05),0_2px_2px_rgba(0,0,0,0.05),0_2px_4px_rgba(0,0,0,0.05)] aria-disabled:pointer-events-none dark:inset-shadow-[0_1px_rgb(255_255_255/0.15)]'
                )}
                href={
                  currentPage === totalPages
                    ? undefined
                    : `#/page/${currentPage + 1}`
                }
                aria-label='Go to next page'
                aria-disabled={currentPage === totalPages ? true : undefined}
                role={currentPage === totalPages ? 'link' : undefined}
              >
                <RiArrowRightSLine
                  className='size-5'
                  size={20}
                  aria-hidden='true'
                />
              </PaginationLink>
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </CardContent>
    </Card>
  )
}
