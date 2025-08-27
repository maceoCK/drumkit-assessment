import { useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  type SortingState, type ColumnDef,
  type ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { ArrowUpDown, ChevronDown, ChevronUp, Plus } from "lucide-react"
import CreateLoadModal from '@/components/CreateLoadModal'

import './App.css'

type Load = {
  externalTMSLoadID: string
  status: string
  customer?: { name?: string }
  createdAt?: string
  pickup?: { city?: string; state?: string }
  consignee?: { city?: string; state?: string }
  phase?: string
  mode?: string
  serviceType?: string
  services?: string[]
  equipment?: string[]
  customerTotalMiles?: number
  marginAmount?: number
  marginValue?: number
}

// App renders a paginated, sortable table of Loads fetched from the backend
// API. It also exposes a Create modal to post new Loads.
function App() {
  const [loads, setLoads] = useState<Load[]>([])
  const [sorting, setSorting] = useState<SortingState>([])
  const [error, setError] = useState<string | null>(null)
  const [start, setStart] = useState(0)
  const [pageSize] = useState(24)
  const [moreAvailable, setMoreAvailable] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  // Server-side filters
  const [filterStatus] = useState('') // Turvo status code (2101/2102) or empty
  const [filterExternalId] = useState('')
  const [filterCreatedFrom] = useState('') // datetime-local
  const [filterUpdatedTo] = useState('')
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])

  function mapSortForAPI(s: SortingState): string | undefined {
    if (!s || s.length === 0) return undefined
    const first = s[0]
    const fieldMap: Record<string, string> = {
      'externalTMSLoadID': 'customId',
      'createdAt': 'createdDate',
      'status': 'status.code',
      'customer.name': 'customer.name',
    }
    const field = fieldMap[first.id] || first.id
    const dir = first.desc ? 'desc' : 'asc'
    return `${field}:${dir}`
  }

  function buildQuery(nextStart = start) {
    const params = new URLSearchParams()
    params.set('start', String(nextStart))
    params.set('pageSize', String(pageSize))
    if (filterStatus) params.set('status[eq]', filterStatus)
    if (filterExternalId) params.set('customId[eq]', filterExternalId)
    if (filterCreatedFrom) params.set('created[gte]', new Date(filterCreatedFrom).toISOString())
    if (filterUpdatedTo) params.set('updated[lte]', new Date(filterUpdatedTo).toISOString())
    const sort = mapSortForAPI(sorting)
    if (sort) params.set('sortBy', sort)
    return params.toString()
  }

  const API_BASE = import.meta.env.VITE_API_BASE?.replace(/\/$/, '') || ''

  async function fetchLoads(nextStart = 0) {
    try {
      setError(null)
      const qs = buildQuery(nextStart)
      const r = await fetch(`${API_BASE}/api/loads?${qs}`)
      if (!r.ok) throw new Error('Failed to fetch loads')
      const data = await r.json()
      const items: Load[] = Array.isArray(data) ? data : (data?.items ?? [])
      setLoads(items)
      const more = !!(data?.pagination?.moreAvailable)
      setMoreAvailable(more)
      setStart(nextStart)
    } catch (e: any) {
      setError(e?.message ?? 'Failed to fetch loads')
    }
  }

  const columns: ColumnDef<Load>[] = useMemo(() => [
    {
      header: ({ column }) => {
        return (
          <div className="space-y-2">
            <Button
              variant="ghost"
              className={column.getIsSorted() ? 'text-primary' : ''}
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            >
              External ID
              {column.getIsSorted() === 'asc' ? (
                <ChevronUp className="ml-2 h-4 w-4" />
              ) : column.getIsSorted() === 'desc' ? (
                <ChevronDown className="ml-2 h-4 w-4" />
              ) : (
                <ArrowUpDown className="ml-2 h-4 w-4" />
              )}
            </Button>
            <Input
              placeholder="Filter"
              value={(column.getFilterValue() as string) ?? ''}
              onChange={(e) => column.setFilterValue(e.target.value)}
              className="h-8"
            />
          </div>
        )
      },
      accessorKey: 'externalTMSLoadID',
      cell: ({ row }) => {
        return <div>{row.original.externalTMSLoadID}</div>
      }
    },
    {
      header: ({ column }) => (
        <div className="space-y-2">
          <Button
            variant="ghost"
            className={column.getIsSorted() ? 'text-primary' : ''}
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          >
            Phase
            {column.getIsSorted() === 'asc' ? (
              <ChevronUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ChevronDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
          <Input
            placeholder="Filter"
            value={(column.getFilterValue() as string) ?? ''}
            onChange={(e) => column.setFilterValue(e.target.value)}
            className="h-8"
          />
        </div>
      ),
      accessorKey: 'phase',
      cell: ({ row }) => (<div>{row.original.phase ?? '-'}</div>)
    },
    {
      header: ({ column }) => (
        <div className="space-y-2">
          <Button
            variant="ghost"
            className={column.getIsSorted() ? 'text-primary' : ''}
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          >
            Mode
            {column.getIsSorted() === 'asc' ? (
              <ChevronUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ChevronDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
          <Input
            placeholder="Filter"
            value={(column.getFilterValue() as string) ?? ''}
            onChange={(e) => column.setFilterValue(e.target.value)}
            className="h-8"
          />
        </div>
      ),
      accessorKey: 'mode',
      cell: ({ row }) => (<div>{row.original.mode ?? '-'}</div>)
    },
    {
      header: ({ column }) => (
        <div className="space-y-2">
          <Button
            variant="ghost"
            className={column.getIsSorted() ? 'text-primary' : ''}
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          >
            Service Type
            {column.getIsSorted() === 'asc' ? (
              <ChevronUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ChevronDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
          <Input
            placeholder="Filter"
            value={(column.getFilterValue() as string) ?? ''}
            onChange={(e) => column.setFilterValue(e.target.value)}
            className="h-8"
          />
        </div>
      ),
      accessorKey: 'serviceType',
      cell: ({ row }) => (<div>{row.original.serviceType ?? '-'}</div>)
    },
    {
      header: ({ column }) => (
        <div className="space-y-2">
          <Button
            variant="ghost"
            className={column.getIsSorted() ? 'text-primary' : ''}
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          >
            Services
            {column.getIsSorted() === 'asc' ? (
              <ChevronUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ChevronDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
          <Input
            placeholder="Filter"
            value={(column.getFilterValue() as string) ?? ''}
            onChange={(e) => column.setFilterValue(e.target.value)}
            className="h-8"
          />
        </div>
      ),
      accessorKey: 'services',
      cell: ({ row }) => {
        const arr = row.original.services || []
        return <div>{arr.length ? arr.join(', ') : '-'}</div>
      }
    },
    {
      header: ({ column }) => (
        <div className="space-y-2">
          <Button
            variant="ghost"
            className={column.getIsSorted() ? 'text-primary' : ''}
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          >
            Equipment
            {column.getIsSorted() === 'asc' ? (
              <ChevronUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ChevronDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
          <Input
            placeholder="Filter"
            value={(column.getFilterValue() as string) ?? ''}
            onChange={(e) => column.setFilterValue(e.target.value)}
            className="h-8"
          />
        </div>
      ),
      accessorKey: 'equipment',
      cell: ({ row }) => {
        const arr = row.original.equipment || []
        return <div>{arr.length ? arr.join(', ') : '-'}</div>
      }
    },
    {
      header: ({ column }) => (
        <div className="space-y-2">
          <Button
            variant="ghost"
            className={column.getIsSorted() ? 'text-primary' : ''}
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          >
            Miles
            {column.getIsSorted() === 'asc' ? (
              <ChevronUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ChevronDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
          <Input
            placeholder="Filter"
            value={(column.getFilterValue() as string) ?? ''}
            onChange={(e) => column.setFilterValue(e.target.value)}
            className="h-8"
          />
        </div>
      ),
      accessorKey: 'customerTotalMiles',
      cell: ({ row }) => (<div>{row.original.customerTotalMiles ?? '-'}</div>)
    },
    {
      header: ({ column }) => (
        <div className="space-y-2">
          <Button
            variant="ghost"
            className={column.getIsSorted() ? 'text-primary' : ''}
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          >
            Margin
            {column.getIsSorted() === 'asc' ? (
              <ChevronUp className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'desc' ? (
              <ChevronDown className="ml-2 h-4 w-4" />
            ) : (
              <ArrowUpDown className="ml-2 h-4 w-4" />
            )}
          </Button>
          <Input
            placeholder="Filter"
            value={(column.getFilterValue() as string) ?? ''}
            onChange={(e) => column.setFilterValue(e.target.value)}
            className="h-8"
          />
        </div>
      ),
      accessorKey: 'marginAmount',
      cell: ({ row }) => (<div>{row.original.marginAmount ?? '-'}</div>)
    },
    {
      header: ({ column }) => {
        return (
          <div className="space-y-2">
            <Button
              variant="ghost"
              className={column.getIsSorted() ? 'text-primary' : ''}
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            >
              Pickup
              {column.getIsSorted() === 'asc' ? (
                <ChevronUp className="ml-2 h-4 w-4" />
              ) : column.getIsSorted() === 'desc' ? (
                <ChevronDown className="ml-2 h-4 w-4" />
              ) : (
                <ArrowUpDown className="ml-2 h-4 w-4" />
              )}
            </Button>
            <Input
              placeholder="Filter"
              value={(column.getFilterValue() as string) ?? ''}
              onChange={(e) => column.setFilterValue(e.target.value)}
              className="h-8"
            />
          </div>
        )
      },
      accessorKey: 'pickup',
      cell: ({ row }) => {
        const p = row.original.pickup
        const text = [p?.city, p?.state].filter(Boolean).join(', ')
        return <div>{text || '-'}</div>
      }
    },
    {
      header: ({ column }) => {
        return (
          <div className="space-y-2">
            <Button
              variant="ghost"
              className={column.getIsSorted() ? 'text-primary' : ''}
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            >
              Destination
              {column.getIsSorted() === 'asc' ? (
                <ChevronUp className="ml-2 h-4 w-4" />
              ) : column.getIsSorted() === 'desc' ? (
                <ChevronDown className="ml-2 h-4 w-4" />
              ) : (
                <ArrowUpDown className="ml-2 h-4 w-4" />
              )}
            </Button>
            <Input
              placeholder="Filter"
              value={(column.getFilterValue() as string) ?? ''}
              onChange={(e) => column.setFilterValue(e.target.value)}
              className="h-8"
            />
          </div>
        )
      },
      accessorKey: 'consignee',
      cell: ({ row }) => {
        const c = row.original.consignee
        const text = [c?.city, c?.state].filter(Boolean).join(', ')
        return <div>{text || '-'}</div>
      }
    },
    {
      header: ({ column }) => {
        return (
          <div className="space-y-2">
            <Button
              variant="ghost"
              className={column.getIsSorted() ? 'text-primary' : ''}
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            >
              Created
              {column.getIsSorted() === 'asc' ? (
                <ChevronUp className="ml-2 h-4 w-4" />
              ) : column.getIsSorted() === 'desc' ? (
                <ChevronDown className="ml-2 h-4 w-4" />
              ) : (
                <ArrowUpDown className="ml-2 h-4 w-4" />
              )}
            </Button>
            <Input
              placeholder="Filter"
              value={(column.getFilterValue() as string) ?? ''}
              onChange={(e) => column.setFilterValue(e.target.value)}
              className="h-8"
            />
          </div>
        )
      },
      accessorKey: 'createdAt',
      cell: ({ row }) => {
        const v = row.original.createdAt
        const d = v ? new Date(v) : null
        return <div>{d ? d.toLocaleString() : '-'}</div>
      }
    },
    
    {
      header: ({ column }) => {
        return (
          <div className="space-y-2">
            <Button
              variant="ghost"
              className={column.getIsSorted() ? 'text-primary' : ''}
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            >
              Customer
              {column.getIsSorted() === 'asc' ? (
                <ChevronUp className="ml-2 h-4 w-4" />
              ) : column.getIsSorted() === 'desc' ? (
                <ChevronDown className="ml-2 h-4 w-4" />
              ) : (
                <ArrowUpDown className="ml-2 h-4 w-4" />
              )}
            </Button>
            <Input
              placeholder="Filter"
              value={(column.getFilterValue() as string) ?? ''}
              onChange={(e) => column.setFilterValue(e.target.value)}
              className="h-8"
            />
          </div>
        )
      },
      accessorKey: 'customer.name',
      cell: ({ row }) => {
        return <div>{row.original.customer?.name ?? '-'}</div>
      }
    },
    
    {
      accessorKey: 'status',
      header: ({ column }) => {
        return (
          <div className="space-y-2">
            <Button
              variant="ghost"
              className={column.getIsSorted() ? 'text-primary' : ''}
              onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
            >
              Status
              {column.getIsSorted() === 'asc' ? (
                <ChevronUp className="ml-2 h-4 w-4" />
              ) : column.getIsSorted() === 'desc' ? (
                <ChevronDown className="ml-2 h-4 w-4" />
              ) : (
                <ArrowUpDown className="ml-2 h-4 w-4" />
              )}
            </Button>
            <Input
              placeholder="Filter"
              value={(column.getFilterValue() as string) ?? ''}
              onChange={(e) => column.setFilterValue(e.target.value)}
              className="h-8"
            />
          </div>
        )
      },
      cell: ({ row }) => {
        return <div><StatusBadge value={row.original.status} /></div>
      }
    },
  ], [])

  const table = useReactTable({
    data: loads,
    columns,
    state: {
      sorting,
      columnFilters,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    manualSorting: true,
  })

  useEffect(() => {
    fetchLoads(0)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Refetch when sorting changes so server applies ordering
  useEffect(() => {
    fetchLoads(0)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(sorting)])

  function StatusBadge({ value }: { value: string }) {
    const v = (value || '').toUpperCase()
    const cls = v === 'COVERED' ? 'bg-green-100 text-green-800' : v === 'NEW' || v === 'TENDERED' ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-800'
    return <span className={`px-2 py-0.5 rounded text-xs font-medium ${cls}`}>{value}</span>
  }

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <CreateLoadModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onSuccess={async () => { await fetchLoads(0) }}
      />

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Loads</CardTitle>
            <Button size="sm" onClick={() => setShowCreate(true)} aria-label="Create Load">
              <Plus className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          {error && (
            <div className="text-sm text-red-600">{error}</div>
          )}

          <div className="flex items-center justify-end gap-3">
            <div className="text-sm text-muted-foreground">Showing {table.getFilteredRowModel().rows.length} items</div>
          </div>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
                        {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows.length > 0 ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow key={row.id}>
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={columns.length} className="text-center text-muted-foreground">No loads</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
          <div className="flex items-center justify-end gap-3">
            <div className="text-sm">Start {start}</div>
            <Button variant="outline" size="sm" onClick={() => fetchLoads(Math.max(0, start - pageSize))} disabled={start === 0}>Prev</Button>
            <Button variant="outline" size="sm" onClick={() => fetchLoads(start + pageSize)} disabled={!moreAvailable}>Next</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default App
