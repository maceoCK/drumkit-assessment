import { useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import './App.css'

type Load = {
  externalTMSLoadID: string
  status: string
  customer?: { name?: string }
}

function App() {
  const [loads, setLoads] = useState<Load[]>([])
  const [externalTMSLoadID, setExternalTMSLoadID] = useState('')
  const [status, setStatus] = useState('NEW')
  const [pickupAt, setPickupAt] = useState<string>('')
  const [deliveryAt, setDeliveryAt] = useState<string>('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('')
  const [start, setStart] = useState(0)
  const [pageSize] = useState(24)
  const [moreAvailable, setMoreAvailable] = useState(false)
  // Server-side filters
  const [filterStatus, setFilterStatus] = useState('') // Turvo status code (2101/2102) or empty
  const [filterExternalId, setFilterExternalId] = useState('')
  const [filterCreatedFrom, setFilterCreatedFrom] = useState('') // datetime-local
  const [filterUpdatedTo, setFilterUpdatedTo] = useState('')

  function buildQuery(nextStart = start) {
    const params = new URLSearchParams()
    params.set('start', String(nextStart))
    params.set('pageSize', String(pageSize))
    if (filterStatus) params.set('status[eq]', filterStatus)
    if (filterExternalId) params.set('customId[eq]', filterExternalId)
    if (filterCreatedFrom) params.set('created[gte]', new Date(filterCreatedFrom).toISOString())
    if (filterUpdatedTo) params.set('updated[lte]', new Date(filterUpdatedTo).toISOString())
    return params.toString()
  }

  async function fetchLoads(nextStart = 0) {
    try {
      setError(null)
      const qs = buildQuery(nextStart)
      const r = await fetch(`/api/loads?${qs}`)
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

  useEffect(() => {
    fetchLoads(0)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const res = await fetch('/api/loads', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          externalTMSLoadID,
          status,
          customer: { name: '' },
          pickup: { name: '', addressLine1: '', city: '', state: '', zipcode: '', country: 'US', readyTime: pickupAt ? new Date(pickupAt).toISOString() : undefined },
          consignee: { name: '', addressLine1: '', city: '', state: '', zipcode: '', country: 'US', mustDeliver: deliveryAt ? new Date(deliveryAt).toISOString() : undefined },
          specifications: {}
        })
      })
      if (!res.ok) throw new Error('Failed to create load')
      setExternalTMSLoadID('')
      await fetchLoads(0)
      setPickupAt('')
      setDeliveryAt('')
    } catch (e: any) {
      setError(e?.message ?? 'Failed to create load')
    } finally {
      setSubmitting(false)
    }
  }

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return loads
    return loads.filter(l =>
      l.externalTMSLoadID.toLowerCase().includes(q) ||
      (l.customer?.name || '').toLowerCase().includes(q) ||
      (l.status || '').toLowerCase().includes(q)
    )
  }, [loads, query])

  function StatusBadge({ value }: { value: string }) {
    const v = (value || '').toUpperCase()
    const cls = v === 'COVERED' ? 'bg-green-100 text-green-800' : v === 'NEW' || v === 'TENDERED' ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-800'
    return <span className={`px-2 py-0.5 rounded text-xs font-medium ${cls}`}>{value}</span>
  }

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Create Load</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="grid gap-4 max-w-2xl grid-cols-1 sm:grid-cols-2">
            <div className="grid gap-2 sm:col-span-2">
              <Label htmlFor="extId">External TMS Load ID</Label>
              <Input id="extId" value={externalTMSLoadID} onChange={e => setExternalTMSLoadID(e.target.value)} required />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="status">Status</Label>
              <Input id="status" value={status} onChange={e => setStatus(e.target.value)} />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="pickupAt">Pickup Time</Label>
              <Input id="pickupAt" type="datetime-local" value={pickupAt} onChange={e => setPickupAt(e.target.value)} />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="deliveryAt">Delivery Time</Label>
              <Input id="deliveryAt" type="datetime-local" value={deliveryAt} onChange={e => setDeliveryAt(e.target.value)} />
            </div>
            <div className="flex gap-2 sm:col-span-2">
              <Button type="submit" disabled={submitting}>{submitting ? 'Creating…' : 'Create'}</Button>
              <Button type="button" variant="outline" onClick={() => fetchLoads(start)}>Refresh</Button>
            </div>
            {error && <div className="text-sm text-red-600 sm:col-span-2">{error}</div>}
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Loads</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="grid sm:grid-cols-5 gap-3 items-end">
            <div className="grid gap-1">
              <Label htmlFor="filterStatus">Status code</Label>
              <Input id="filterStatus" placeholder="e.g. 2101" value={filterStatus} onChange={e => setFilterStatus(e.target.value)} />
            </div>
            <div className="grid gap-1">
              <Label htmlFor="filterExternal">External ID</Label>
              <Input id="filterExternal" value={filterExternalId} onChange={e => setFilterExternalId(e.target.value)} />
            </div>
            <div className="grid gap-1">
              <Label htmlFor="createdFrom">Created from</Label>
              <Input id="createdFrom" type="datetime-local" value={filterCreatedFrom} onChange={e => setFilterCreatedFrom(e.target.value)} />
            </div>
            <div className="grid gap-1">
              <Label htmlFor="updatedTo">Updated to</Label>
              <Input id="updatedTo" type="datetime-local" value={filterUpdatedTo} onChange={e => setFilterUpdatedTo(e.target.value)} />
            </div>
            <div className="flex gap-2">
              <Button type="button" onClick={() => fetchLoads(0)}>Apply</Button>
              <Button type="button" variant="outline" onClick={() => { setFilterStatus(''); setFilterExternalId(''); setFilterCreatedFrom(''); setFilterUpdatedTo(''); fetchLoads(0) }}>Clear</Button>
            </div>
          </div>

          <div className="flex items-center justify-between gap-3">
            <Input placeholder="Search this page" value={query} onChange={e => setQuery(e.target.value)} />
            <div className="text-sm text-muted-foreground">Showing {filtered.length} items</div>
          </div>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>External ID</TableHead>
                  <TableHead>Customer</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((l, idx) => (
                  <TableRow key={idx}>
                    <TableCell className="font-medium">{l.externalTMSLoadID}</TableCell>
                    <TableCell>{l.customer?.name ?? '-'}</TableCell>
                    <TableCell><StatusBadge value={l.status} /></TableCell>
                  </TableRow>
                ))}
                {filtered.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={3} className="text-center text-muted-foreground">No loads</TableCell>
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
