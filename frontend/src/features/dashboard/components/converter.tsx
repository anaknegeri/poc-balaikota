'use client'

import { RiArrowDownLine, RiSettings4Line } from '@remixicon/react'
import { I18nProvider, Input, Label, NumberField } from 'react-aria-components'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

interface ConverterFieldProps {
  className?: string
  isLast?: boolean
  defaultValue: number
  balance: string
  defaultCoin: string
  coins: {
    id: string
    name: string
    image: string
  }[]
}

function ConverterField({
  className,
  isLast,
  defaultValue,
  balance,
  defaultCoin,
  coins,
}: ConverterFieldProps) {
  return (
    <>
      {isLast && (
        <div
          className='from-primary to-primary-to absolute top-1/2 flex size-10 -translate-y-1/2 items-center justify-center rounded-full bg-linear-to-b inset-shadow-[0_1px_rgb(255_255_255/0.15)]'
          aria-hidden='true'
        >
          <RiArrowDownLine className='text-primary-foreground' size={20} />
        </div>
      )}
      <Card
        className={cn(
          'dark:bg-card/64 relative w-full flex-row items-center justify-between gap-2 p-5',
          isLast
            ? '[mask-image:radial-gradient(ellipse_26px_24px_at_50%_0%,transparent_0,_transparent_24px,_black_25px)]'
            : '[mask-image:radial-gradient(ellipse_26px_24px_at_50%_100%,transparent_0,_transparent_24px,_black_25px)]',
          className
        )}
      >
        {isLast && (
          <div
            className='absolute -top-px left-1/2 h-[25px] w-[50px] -translate-x-1/2 rounded-b-full border-x border-b border-white/15'
            aria-hidden='true'
          ></div>
        )}
        <div className='grow'>
          <I18nProvider locale='en-US'>
            <NumberField
              defaultValue={defaultValue}
              minValue={0}
              formatOptions={{
                minimumFractionDigits: 1,
                maximumFractionDigits: 2,
                useGrouping: true,
              }}
            >
              <Label className='sr-only'>Amount</Label>
              <Input className='focus:bg-card/64 mb-0.5 -ml-1 w-full max-w-40 appearance-none rounded-lg bg-transparent px-1 py-0.5 text-2xl font-semibold focus-visible:outline-none' />
            </NumberField>
          </I18nProvider>
          <div className='text-muted-foreground text-xs'>
            <span className='text-muted-foreground/70'>Balance: </span>
            {balance}
          </div>
        </div>
        <div>
          <Select defaultValue={defaultCoin}>
            <SelectTrigger className='[&>span_svg]:text-muted-foreground/80 bg-card/64 hover:bg-card/80 h-8 rounded-full border-0 p-1 pr-2 shadow-lg inset-shadow-[0_1px_rgb(255_255_255/0.15)] [&>span]:flex [&>span]:items-center [&>span]:gap-2 [&>span_svg]:shrink-0'>
              <SelectValue placeholder='Select coin' />
            </SelectTrigger>
            <SelectContent
              className='dark [&_*[role=option]>span>svg]:text-muted-foreground/80 border-none bg-zinc-800 inset-shadow-[0_1px_rgb(255_255_255/0.15)] shadow-black/10 [&_*[role=option]]:ps-2 [&_*[role=option]]:pe-8 [&_*[role=option]>span]:start-auto [&_*[role=option]>span]:end-2 [&_*[role=option]>span]:flex [&_*[role=option]>span]:items-center [&_*[role=option]>span]:gap-2 [&_*[role=option]>span>svg]:shrink-0'
              align='center'
            >
              {coins.map((coin) => (
                <SelectItem key={coin.id} value={coin.id}>
                  <img
                    className='shrink-0 rounded-full shadow-[0px_0px_0px_1px_rgba(0,0,0,0.04),0_1px_1px_rgba(0,0,0,0.05),0_2px_2px_rgba(0,0,0,0.05),0_2px_4px_rgba(0,0,0,0.05)]'
                    src={coin.image}
                    width={24}
                    height={24}
                    alt={coin.name}
                  />
                  <span className='truncate text-xs font-medium uppercase'>
                    {coin.name}
                  </span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </Card>
    </>
  )
}

export function Converter() {
  const coins = [
    {
      id: '1',
      name: 'Ark',
      image:
        'https://raw.githubusercontent.com/origin-space/origin-images/refs/heads/main/exp4/coin-01-sm-dark_hkrvvm.svg',
    },
    {
      id: '2',
      name: 'Tok',
      image:
        'https://raw.githubusercontent.com/origin-space/origin-images/refs/heads/main/exp4/coin-02-sm-dark_iqldgv.svg',
    },
  ]

  function ConverterContent() {
    return (
      <>
        <div className='relative mb-4 flex flex-col items-center gap-1'>
          <ConverterField
            defaultValue={15494.9}
            balance='24,579'
            defaultCoin='2'
            coins={coins}
          />
          <ConverterField
            isLast
            defaultValue={12984.2}
            balance='1,379.2'
            defaultCoin='1'
            coins={coins}
          />
        </div>
        <div className='text-muted-foreground/50 mb-2 ps-3 text-xs font-medium uppercase'>
          Summary
        </div>
        <Card className='gap-0 rounded-[0.75rem] p-4'>
          <ul className='text-sm'>
            <li className='border-card/50 mb-3 flex items-center justify-between border-b pb-3'>
              <span className='text-muted-foreground'>Transaction Value</span>
              <span className='font-medium'>$2,867</span>
            </li>
            <li className='border-card/50 mb-3 flex items-center justify-between border-b pb-3'>
              <span className='text-muted-foreground'>Network Fees</span>
              <span className='font-medium'>$31.2</span>
            </li>
            <li className='border-card/50 mb-3 flex items-center justify-between border-b pb-3'>
              <span className='text-muted-foreground'>Order Net</span>
              <span className='font-medium'>$2,898.2</span>
            </li>
          </ul>
          <Button size='lg' className='w-full'>
            Confirm
          </Button>
          <div className='mt-3 text-center text-xs uppercase'>
            1 <span className='text-muted-foreground'>ARK =</span> 1,574.04{' '}
            <span className='text-muted-foreground'>TOK</span>
          </div>
        </Card>
      </>
    )
  }

  return (
    <Tabs defaultValue='tab-1' className='flex-1 gap-5'>
      <div className='flex items-center gap-2'>
        <TabsList className='bg-background dark:bg-card/64 flex w-full p-0 shadow-md *:not-first:ms-px dark:inset-shadow-[0_1px_rgb(255_255_255/0.15)]'>
          <TabsTrigger
            value='tab-1'
            className='before:bg-border dark:before:bg-card relative flex-1 before:absolute before:inset-y-2 before:-left-px before:w-px first:before:hidden data-[state=active]:bg-transparent data-[state=active]:shadow-none'
          >
            Convert
          </TabsTrigger>
          <TabsTrigger
            value='tab-2'
            className='before:bg-border dark:before:bg-card relative flex-1 before:absolute before:inset-y-2 before:-left-px before:w-px first:before:hidden data-[state=active]:bg-transparent data-[state=active]:shadow-none'
          >
            Buy
          </TabsTrigger>
          <TabsTrigger
            value='tab-3'
            className='before:bg-border dark:before:bg-card relative flex-1 before:absolute before:inset-y-2 before:-left-px before:w-px first:before:hidden data-[state=active]:bg-transparent data-[state=active]:shadow-none'
          >
            Send
          </TabsTrigger>
        </TabsList>
        <Button
          size='icon'
          variant='ghost'
          className='text-muted-foreground hover:text-foreground/80 size-8 shrink-0'
        >
          <span className='sr-only'>Settings</span>
          <RiSettings4Line className='size-5' size={20} aria-hidden='true' />
        </Button>
      </div>
      <div className='dark bg-background dark:bg-secondary/64 rounded-2xl p-2'>
        <TabsContent value='tab-1'>
          <ConverterContent />
        </TabsContent>
        <TabsContent value='tab-2'>
          <ConverterContent />
        </TabsContent>
        <TabsContent value='tab-3'>
          <ConverterContent />
        </TabsContent>
      </div>
    </Tabs>
  )
}
